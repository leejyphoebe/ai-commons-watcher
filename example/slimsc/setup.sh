#!/bin/bash

dir="/scratch/users/ntu/$USER"
setup_dir="$dir/.ai-commons/setup"
sif_dir="$dir/sif_images"

# explicitly create the logs directory
mkdir -p "logs" "$dir/slimsc_logs" "$dir/slimsc_results" "$sif_dir"

# pull singularity images
qsub $setup_dir/pull_vllm_sif.pbs $1
echo "singularity pull vllm image status code: $?"

qsub $setup_dir/pull_slimsc_sif.pbs $1
echo "singularity pull slimsc image status code: $?"

qsub $setup_dir/download_r1.pbs
echo "huggingface download model status code: $?"
