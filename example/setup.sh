# #!/bin/bash

# repo="git@github.com:googlercolin/slimsc.git"
# dir="/scratch/users/ntu/$USER"
# slimsc_dir="$dir/slimsc"
# setup_dir="$dir/.ai-commons/setup"
# sif_dir="$dir/sif_images"

# if ! git clone "${repo}" "${slimsc_dir}" 2>/dev/null && [ -d "${slimsc_dir}" ] ; then
#     echo "Clone failed because the folder ${slimsc_dir} exists"
#     cd "${slimsc_dir}" && git pull
# fi

# # explicitly create the logs directory
# mkdir -p "logs" "$dir/slimsc_logs" "$dir/slimsc_results" "$sif_dir"

# # pull singularity images
# module load singularity git python/3.10.9
# rm $sif_dir/vllm.sif
# rm $sif_dir/slimsc.sif
# export $(cat $setup_dir/dockerhub.env | xargs)
# export SINGULARITY_CACHEDIR=$dir/.singularity
# singularity pull $sif_dir/vllm.sif docker://broccolin/vllm-nscc-edit:latest
# singularity pull $sif_dir/slimsc.sif docker://meowth16/anyrepo:slimsc
# # qsub $setup_dir/download_r1.pbs

# cd $slimsc_dir && git checkout hmmt_parallel_test
