#!/bin/bash

# Parâmetros
SCENARIO=$1
DURATION=$2
INTERVAL=10
NUM_SAMPLES=$((DURATION / INTERVAL))
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Verificar espaço em disco antes de iniciar
DISK_USAGE=$(df -h / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 90 ]; then
 echo "Erro: Espaço em disco insuficiente ($DISK_USAGE% usado). Limpando logs antigos..."
 find ~/metrics_* -mtime +1 -delete
fi

# Arquivos de log
LOG_DIR="metrics_${TIMESTAMP}"
mkdir -p $LOG_DIR

CPU_LOG="${LOG_DIR}/cpu_${SCENARIO}.log"
MEMORY_LOG="${LOG_DIR}/memory_${SCENARIO}.log"
NETWORK_LOG="${LOG_DIR}/network_${SCENARIO}.log"

# Função para coletar dados de CPU
collect_cpu() {
 echo "Coletando dados de CPU..."
 sar -u ALL $INTERVAL $NUM_SAMPLES > "$CPU_LOG"
}

# Função para coletar dados de memória
collect_memory() {
 echo "Coletando dados de memória..."
 sar -r $INTERVAL $NUM_SAMPLES > "$MEMORY_LOG"
}

# Função para coletar dados de rede
collect_network() {
 echo "Coletando dados de largura de banda..."
 sar -n DEV $INTERVAL $NUM_SAMPLES > "$NETWORK_LOG"
}

# Coleta de dados em paralelo
collect_cpu &
collect_memory &
collect_network

# Espera todas as funções de coleta terminarem
wait

echo "Coleta de dados concluída para o cenário $SCENARIO."
echo "Logs salvos em $LOG_DIR"

# Compactar logs para facilitar transferência
tar -czf "metrics_${SCENARIO}.tar.gz" $LOG_DIR

# Remover diretório original após compactação para economizar espaço
rm -rf $LOG_DIR