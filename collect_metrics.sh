#!/bin/bash

# Parâmetros
SCENARIO=$1
DURATION=$2
INTERVAL=10
NUM_SAMPLES=$((DURATION / INTERVAL))
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

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