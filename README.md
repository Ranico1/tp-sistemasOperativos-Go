Deploy
Clonar el tp en /home/utnso

Cambiar las ips de los modulos de la siguiente forma: ./modificarIp MODULO_QUE_VOY_A_EJECUTAR CLAVE "VALOR" Ejemplo: ./modificarIp memoria ip_kernel "192.0.0.7"

Segun la prueba que quiera ejecutar, buildear el modulo con el make

Prueba Planificacion
Filesystem : make filesystem ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/filesystem/config/config_planificacion.json

Memoria : make memoria ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/memoria/config/config_planificacion.json

CPU : make cpu ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/cpu/config/config.json

Kernel:

FIFO: make kernel ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/kernel/config/config_planificacionFifo.json COD=PLANI_PROC TAM="32"
PRIORIDADES:make kernel ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/kernel/config/config_planificacionPrio.json COD=PLANI_PROC TAM="32"
CMN: make kernel ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/kernel/config/config_planificacionCmn.json COD=PLANI_PROC TAM="32"
Prueba Race Condition
Filesystem : make filesystem ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/filesystem/config/config_raceCondition.json

Memoria : make memoria ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/memoria/config/config_raceCondition.json

CPU : make cpu ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/cpu/config/config.json

Kernel:

Quantum 750: make kernel ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/kernel/config/config_raceCondition1.json COD=RECURSOS_MUTEX_PROC TAM="32"
Quantum 150: make kernel ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/kernel/config/config_raceCondition2.json COD=RECURSOS_MUTEX_PROC TAM="32"
Prueba Particiones Fijas
Filesystem : make filesystem ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/filesystem/config/config_partFijas.json

CPU : make cpu ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/cpu/config/config.json

Memoria:

FIRST: make memoria ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/memoria/config/config_partFijasFirst.json
BEST: make memoria ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/memoria/config/config_partFijasBest.json
WORST: make memoria ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/memoria/config/config_partFijasWorst.json
Kernel : make kernel ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/kernel/config/config_partFijas.json COD=MEM_FIJA_BASE TAM="12"

Prueba Particiones Dinamicas
Filesystem : make filesystem ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/filesystem/config/config_partDinamicas.json

CPU : make cpu ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/cpu/config/config.json

Memoria : make memoria ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/memoria/config/config_partDinamicas.json

Kernel : make kernel ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/kernel/config/config_partDinamicas.json COD=MEM_DINAMICA_BASE TAM="128"

Prueba Fibonacci Sequence
Filesystem : make filesystem ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/filesystem/config/config_fibonacci.json

CPU : make cpu ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/cpu/config/config.json

Memoria : make memoria ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/memoria/config/config_fibonacci.json

Kernel : make kernel ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/kernel/config/config_fibonacci.json COD=PRUEBA_FS TAM="8"

Prueba Stress
Filesystem : make filesystem ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/filesystem/config/config_stress.json

CPU : make cpu ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/cpu/config/config.json

Memoria : make memoria ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/memoria/config/config_stress.json

Kernel : make kernel ENV=prod C=/home/utnso/tp-2024-2c-Futbol-y-Negocios/kernel/config/config_stress.json COD=THE_EMPTINESS_MACHINE TAM="16"

Checkpoint
Para cada checkpoint de control obligatorio, se debe crear un tag en el repositorio con el siguiente formato:

checkpoint-{número}
Donde {número} es el número del checkpoint.

Para crear un tag y subirlo al repositorio, podemos utilizar los siguientes comandos:

git tag -a checkpoint-{número} -m "Checkpoint {número}"
git push origin checkpoint-{número}
