#!/bin/bash

scratch="/scratch/users/ntu/$USER"
module add singularity
uuid=$(uuidgen)

export SINGULARITYENV_SLIMSC_OUTPUT_DIR="$scratch/slimsc_results"
export SINGULARITYENV_SLIMSC_LOGS_DIR="$scratch/slimsc_logs"
singularity exec \
    --no-home \
    --home /slimsc/prune \
    -B $scratch/.ai-commons/experiments/slimsc:/experiments \
    -B $scratch/slimsc_logs:/slimsc/prune/jobs/logs \
    -B $scratch:$scratch \
    $scratch/sif_images/slimsc.sif \
    python /slimsc/prune/jobs/generate_jobs_sif.py \
        --config $1 \
        --job_uuid $uuid \

xargs -a "$scratch/slimsc_logs/${uuid}_pbs_scripts.txt" -I{} echo "qsub {}"
xargs -a "$scratch/slimsc_logs/${uuid}_pbs_scripts.txt" -I{} qsub {}
qstat -a
