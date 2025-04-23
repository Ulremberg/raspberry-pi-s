#!/bin/bash

# Script para execução de testes por 3 dias em Raspberry Pi
# Uso: ./run_long_test.sh [device_ip] [device_name]
# Exemplo: ./run_long_test.sh 192.168.1.100 zero_w

# Verificar argumentos
if [ $# -lt 2 ]; then
 echo "Uso: $0 [device_ip] [device_name]"
 echo "Exemplo: $0 192.168.1.100 zero_w"
 exit 1
fi

DEVICE_IP=$1
DEVICE_NAME=$2

# Validar nome do dispositivo
if [ "$DEVICE_NAME" != "zero_w" ] && [ "$DEVICE_NAME" != "zero_2w" ]; then
 echo "Erro: O nome do dispositivo deve ser 'zero_w' ou 'zero_2w'"
 exit 1
fi

# Diretório para armazenar logs
LOG_DIR="logs_${DEVICE_NAME}_3days_$(date +%Y%m%d_%H%M%S)"
mkdir -p $LOG_DIR

# Definir duração total do teste (3 dias em segundos)
TOTAL_DURATION=$((3 * 24 * 60 * 60))  # 3 dias = 259.200 segundos
START_TIME=$(date +%s)
END_TIME=$((START_TIME + TOTAL_DURATION))

# Função para executar um cenário de teste
run_test() {
 local num_sensors=$1
 local payload=$2
 local frequency=$3
 local duration=$4
 local test_id=$5
 
 echo "========================================================"
 echo "Iniciando teste #$test_id em $DEVICE_NAME ($DEVICE_IP)"
 echo "Sensores: $num_sensors, Payload: $payload, Frequência: $frequency ms"
 echo "Duração: $duration segundos"
 echo "Data/Hora: $(date)"
 echo "========================================================"
 
 # Iniciar coleta de métricas
 echo "Iniciando coleta de métricas..."
 ssh pi@$DEVICE_IP "nohup ./collect_metrics.sh ${DEVICE_NAME}_${test_id}_${num_sensors}_${payload}_${frequency} $duration > /dev/null 2>&1 &"
 sleep 5
 
 # Executar simulador
 echo "Iniciando simulação de sensores..."
 go run simulator.go --device=$DEVICE_NAME --sensors=$num_sensors --payload=$payload --frequency=$frequency --duration=$duration
 
 # Aguardar finalização da coleta de métricas
 echo "Aguardando finalização da coleta de métricas..."
 sleep $((duration + 30))
 
 # Copiar logs de métricas
 echo "Copiando logs de métricas..."
 scp pi@$DEVICE_IP:~/metrics_*${DEVICE_NAME}_${test_id}_${num_sensors}_${payload}_${frequency}*.tar.gz $LOG_DIR/
 
 echo "Teste #$test_id concluído: $num_sensors sensores, payload $payload"
 echo "========================================================"
}

# Verificar se o dispositivo está acessível
echo "Verificando conectividade com $DEVICE_NAME ($DEVICE_IP)..."
ping -c 1 $DEVICE_IP > /dev/null || { echo "Erro: Não foi possível conectar ao dispositivo $DEVICE_IP"; exit 1; }

# Verificar se o script de coleta de métricas está presente no dispositivo
echo "Verificando script de coleta de métricas..."
ssh pi@$DEVICE_IP "test -f ./collect_metrics.sh" || { echo "Erro: Script collect_metrics.sh não encontrado no dispositivo"; exit 1; }

# Iniciar loop de testes
echo "Iniciando bateria de testes em $DEVICE_NAME..."
echo "Início: $(date)"
echo "Término previsto: $(date -d @$END_TIME)"

# Contador de testes
TEST_COUNT=1

# Loop principal - executa até atingir o tempo total
current_time=$(date +%s)
while [ $current_time -lt $END_TIME ]; do
 # Calcular o tempo decorrido em horas
 elapsed_hours=$(( (current_time - START_TIME) / 3600 ))
 day_hour=$((elapsed_hours % 24))
 
 # Variar cenários de teste com base na hora do dia
 if [ $day_hour -ge 8 ] && [ $day_hour -lt 18 ]; then
     # Horário comercial (8h-18h): carga alta
     case $((TEST_COUNT % 3)) in
         0)
             run_test 500 "small" 100 1800 $TEST_COUNT  # 30 minutos
             ;;
         1)
             run_test 300 "medium" 200 1800 $TEST_COUNT  # 30 minutos
             ;;
         2)
             run_test 100 "large" 500 1800 $TEST_COUNT  # 30 minutos
             ;;
     esac
 elif [ $day_hour -ge 18 ] && [ $day_hour -lt 22 ]; then
     # Início da noite (18h-22h): carga média
     run_test 200 "medium" 300 1800 $TEST_COUNT  # 30 minutos
 else
     # Noite/madrugada (22h-8h): carga baixa
     run_test 50 "small" 1000 1800 $TEST_COUNT  # 30 minutos
 fi
 
 # Incrementar contador de testes
 TEST_COUNT=$((TEST_COUNT + 1))
 
 # Verificar espaço em disco e limpar se necessário
 DISK_USAGE=$(ssh pi@$DEVICE_IP "df -h / | tail -1 | awk '{print \$5}' | sed 's/%//'")
 if [ $DISK_USAGE -gt 80 ]; then
     echo "Aviso: Uso de disco alto ($DISK_USAGE%). Limpando arquivos temporários..."
     ssh pi@$DEVICE_IP "find /tmp -type f -mtime +1 -delete"
     ssh pi@$DEVICE_IP "find ~/metrics_* -mtime +1 -delete"
 fi
 
 # Atualizar tempo atual
 current_time=$(date +%s)
 
 # Calcular tempo restante
 remaining_time=$((END_TIME - current_time))
 remaining_hours=$((remaining_time / 3600))
 remaining_minutes=$(( (remaining_time % 3600) / 60 ))
 
 echo "Tempo restante: $remaining_hours horas e $remaining_minutes minutos"
 echo "Testes concluídos até agora: $TEST_COUNT"
 
 # Pausa curta entre testes
 sleep 60
done

echo "Todos os testes foram concluídos para $DEVICE_NAME."
echo "Total de testes executados: $TEST_COUNT"
echo "Os logs estão disponíveis em: $LOG_DIR"

# Extrair arquivos tar.gz
echo "Extraindo logs para análise..."
mkdir -p $LOG_DIR/extracted
for tarfile in $LOG_DIR/*.tar.gz; do
 tar -xzf $tarfile -C $LOG_DIR/extracted
done

echo "Testes concluídos. Você pode analisar os resultados em $LOG_DIR/extracted"