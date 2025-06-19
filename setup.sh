#!/bin/bash

rm /scratch/users/ntu/$USER/sif_images/save_model.sif
singularity pull /scratch/users/ntu/$USER/sif_images/vllm_nscc.sif docker://broccolin/vllm-nscc:latest
git clone https://github.com/googlercolin/slimsc.git /scratch/users/ntu/$USER/slimsc
