#!/bin/bash

# Script para execução de testes de longa duração (3 dias)
TOTAL_DURATION=259200  # 3 dias em segundos
START_TIME=$(date +%s)
END_TIME=$((START_TIME + TOTAL_DURATION))

# Diretório para armazenar logs
LOG_DIR="logs_$(date +%Y%m%d_%H%M%S)"
mkdir -p $LOG_DIR

# Função para executar um cenário específico
run_scenario() {
local scenario=$1
local sensors=$2
local payload=$3
local frequency=$4
local duration=$5
local device=$6  # "zero_w" ou "zero_2w"

echo "Executando cenário: $scenario em $device com $sensors sensores, payload $payload, frequência $frequency ms, duração $duration s"

# Iniciar coleta de métricas
ssh pi@${device} "nohup ./collect_metrics.sh ${scenario}_${sensors}_${payload}_${frequency} $duration > /dev/null 2>&1 &"

# Executar simulador com os parâmetros especificados
go run simulator.go --device=${device} --scenario=${scenario} --sensors=${sensors} --payload=${payload} --frequency=${frequency} --duration=${duration}

# Aguardar finalização da coleta de métricas
sleep $((duration + 30))

# Copiar logs de métricas
scp pi@${device}:~/metrics_${scenario}_${sensors}_${payload}_${frequency}*.log $LOG_DIR/

echo "Cenário $scenario concluído"
}

# Definir IPs dos dispositivos
ZERO_W_IP="192.168.1.100"
ZERO_2W_IP="192.168.1.101"

# Loop principal que executa até o fim do período de 3 dias
current_time=$(date +%s)
while [ $current_time -lt $END_TIME ]; do
# Calcular o tempo decorrido em horas
elapsed_hours=$(( (current_time - START_TIME) / 3600 ))
day_hour=$((elapsed_hours % 24))

# Simular diferentes padrões de carga baseados na hora do dia
if [ $day_hour -ge 8 ] && [ $day_hour -lt 18 ]; then
    # Horário comercial (8h-18h): carga alta
    for device in "zero_w" "zero_2w"; do
        if [ "$device" == "zero_w" ]; then
            IP=$ZERO_W_IP
        else
            IP=$ZERO_2W_IP
        fi
        
        # Alternar entre diferentes cenários de alta carga
        case $((elapsed_hours % 3)) in
            0)
                run_scenario "high_load_small" 500 "small" 100 1800 $IP
                ;;
            1)
                run_scenario "high_load_medium" 300 "medium" 200 1800 $IP
                ;;
            2)
                run_scenario "high_load_large" 100 "large" 500 1800 $IP
                ;;
        esac
    done
elif [ $day_hour -ge 18 ] && [ $day_hour -lt 22 ]; then
    # Início da noite (18h-22h): carga média
    for device in "zero_w" "zero_2w"; do
        if [ "$device" == "zero_w" ]; then
            IP=$ZERO_W_IP
        else
            IP=$ZERO_2W_IP
        fi
        
        run_scenario "medium_load" 200 "medium" 300 1800 $IP
    done
else
    # Noite/madrugada (22h-8h): carga baixa
    for device in "zero_w" "zero_2w"; do
        if [ "$device" == "zero_w" ]; then
            IP=$ZERO_W_IP
        else
            IP=$ZERO_2W_IP
        fi
        
        run_scenario "low_load" 50 "small" 1000 1800 $IP
    done
fi

# Atualizar tempo atual
current_time=$(date +%s)
done

echo "Teste de 3 dias concluído. Todos os logs estão em $LOG_DIR"