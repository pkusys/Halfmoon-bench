Benchmark workloads of Halfmoon
==================================

This repository includes the artifacts of our SOSP '23 paper.

### Structure of this repository ###

* `dockerfiles`: Dockerfiles for building relevant Docker images.
* `workloads`: source code of Halfmoon client library and evalution workloads. 
* `experiments`: scripts for running individual experiments.
* `scripts`: helper scripts for setting up AWS EC2 environment, building Docker images, and summarizing experiment results.
* `halfmoon`: git submodule containing our implementation of [Halfmoon](https://github.com/pkusys/Halfmoon)'s logging layer, which is based on SOSP '21 paper [Boki](https://github.com/ut-osa/boki)

### Hardware and software dependencies ###

Our evaluation workloads run on AWS EC2 instances in `ap-southeast-1` region.

EC2 VMs for running experiments use a public AMI (`ami-0cfb31bd9b06138c7`), which runs Ubuntu 20.04 with necessary dependencies installed.

### Environment setup ###

#### Setting up the controller machine ####

A controller machine in AWS `ap-southeast-1` region is required for running evaluation scripts. This machine provisions EC2 spot instances that run the actual experiments and does not participate itself.
Therefore the controller can use very small EC2 instance type.
There are two ways to setup the controller machine:

1. Using the EC2 instance (a `t2.micro`) provided by us. Please send us your public key and we will enable SSH login for you.

2. Setting up your own controller using our public controller AMI (`ami-0e9f4c3294198a422`). This AMI has AWS CLI version 1 installed. To grant the controller access to AWS resources (IAM, EC2, etc.), you need to configure AWS CLI with your access key that has necessary permissions (see this [documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)).

After setting up the controller, clone this repository with all git submodules
```
git clone --recursive https://github.com/pkusys/Halfmoon-bench.git
```

Finally, execute `scripts/setup_sshkey.sh` to setup SSH keys that will be used to access experiment VMs.
Please read the notice in `scripts/setup_sshkey.sh` before executing it to see if this script works for your setup.

#### Setting up EC2 security group and placement group ####

Our evaluation workloads run on exactly the same environment as [Boki](https://github.com/ut-osa/boki-benchmarks). The experiment cluster needs to be provisioned with security group `boki` and placement group `boki-experiments`.
The security group includes firewall rules for experiment VMs (including allowing the controller machine to SSH into them),
while the placement group instructs AWS to place experiment VMs close together.

Executing `scripts/aws_provision.sh` on the controller machine creates these groups with correct configurations.

**NOTE**: reviewers using our provided instance may skip this step.

#### (Optional) Building Docker images ####
We also provide the script (`scripts/build_images.sh`) for building relevant Docker images.
The images required for artifact evaluation are already available on DockerHub.
There is no need to rebuild Docker images unless you wish to modify source code of Halfmoon's logging layer (in `halfmoon` directory) or the client library and evaluation workloads (in `workloads` directory).

### Experiment workflow ###

Each directory within `experiments` corresponds to a subsection in our paper's evalution section. 

- **singleop**: microbenchmarks (section 6.1)
- **workflow**: application workloads (section 6.2)
- **overhead**: system overhead (section 6.3)
- **switching**: switching delay (section 6.4)

Each experiment's directory may be further divided into several subdirectories, each corresponding to an individual baseline. The structure of each baseline's directory is as follows:
- `docker-compose.yml` describes configuration of the individual services that make up the evaluation workload.
- `config.json` describes machine configuration and placement of services.
- `nightcore_config.json` describes the worker configuration for each service.
- `run_once.sh` runs the workload once with the configured options. Each run would create a folder in the `results` directory to store the results. The naming of that folder reflects the configuration of this run.
- `run_all.sh` enumerates workload-specific options and executes `run_once.sh` multiple times. This is the entry point for running experiments.

Before executing `run_once.sh`, `run_all.sh` first does VM provisioning by calling `scripts/exp_helper` with sub-command `start-machines`.
After EC2 instances are up, the `run_once.sh` script calls `docker stack deploy` to create the services specified in `docker-compose.yml` in Docker [swarm](https://docs.docker.com/engine/swarm/) mode. It then `ssh`s to a client machine specified in `config.json` to start generating workload requests.

Each experiment's directory contains a `run_quick.sh` that executes all the `run_all.sh` scripts in its subdirectories (except for `experiments/switching` that contains only a single `run_all.sh`). This script covers all relevant experiments to reproduce a figure in the respective sections of the paper. The `run_quick.sh` script accepts a `RUN` parameter to number each execution of the workloads and passes this parameter to `run_all.sh` scripts. For example, `./run_quick.sh 1` would create folders in the `results` directories labeled with the suffix `_1`.

At the top level, `experiments/run.sh` is a push-button script for running all experiments. Reviewers can comment out specific lines and modifies each `RUN` parameter to select or re-run a subset of experiments.

### Estimated Duration ###

- **singleop**: Each baseline runs for about 10 minutes. With four baselines, the total duration should be around 40 minutes.
- **workflow**: Each data point on the figure takes around 5 minutes. Producing all points could be time-consuming. Therefore we only run a subset of the full parameter combination (see the respective `run_all.sh`). The two application workloads take around 2 hours in total.
- **overhead**: Each data point takes around 10 minutes, and we also run a subset of the full parameter combination. The total duration is around 1.5 hours.
- **switching**: Should completes in 5 minutes.


### Visualization ###

Each `run_quick.sh` (or `run_all.sh` in the `experiments/switching`) ends with a call to a python script that summarizes and visualizes the results. The produced figures are stored in `figures` in the respective experiment directory. Like `results`, the naming of folders within `figures` also reflects the `RUN` parameter passed to `run_quick.sh`.

- **singleop**: `microbenchmarks.png` should match Figure 10.
- **workflow**: `hotel.png` and `movie.png` should match Figure 11 and 12, respectively.
- **overhead**: `runtime_overhead_*.png` and `storage_overhead_*.png` should match Figure 13 and 14, respectively. As this experiment is time consuming, we only reproduce the leftmost subfigure in each figure by default. The asterisk in the figure names indicates the selected parameter combination.
- **switching**: `switching.png` should match Figure 15.

### Troubleshooting ###

The experiment workflow may fail due to EC2 or DynamoDB provisioning errors. These errors are transient and can be usually fixed by re-running the experiment. The following workarounds might also be useful.

- The `scripts/exp_helper` may fail to start the EC2 instances with an error message saying "spot instance request not fulfilled after xxx seconds". Please wait a few minutes and retry. Also consider changing the AVAILABILITY_ZONE variable in `scripts/exp_helper` to another area (1a, 1b, or 1c). By default our scripts would skip the experiment if the error happens. Note that other instances may have successfully started. Please contact us to remove these orphaned instances.

- After EC2 instances are up, the experiments may still fail because some instances becomes unreachable from the controller machine. If there is an `machines.json` file in the directory (the one containing `config.json`), please run `scripts/exp_helper` with sub-command `stop-machines` to terminate the existing instances. Before retrying, please manually delete the result of this run in `results`.

- Our experiments provision and populate DynamoDB tables. During this process, there could be an error message saying something like "resource not found". This should only affect a single execution of `run_once.sh`. Please wait till the current script finishes and re-run the affected experiments. Like in the previous case, please manually delete the results of the affected experiments. Our scripts will skip an experiment if its corresponding foler exists in `results`.

Please contact us if any of these errors persists or if there is an unidentified issue.

### License ###

* The logging layer of [Halfmoon](https://github.com/pkusys/Halfmoon) is based on based on [Boki](https://github.com/ut-osa/boki). Halfmoon is licensed under Apache License 2.0, in accordance with Boki.
* The Halfmoon client library and evaluation workloads (`workloads/workflow`) are based on [Beldi codebase](https://github.com/eniac/Beldi) and [BokiFlow](https://github.com/ut-osa/boki-benchmarks). Both are licensed under MIT License, and so is our source code.
* All other source code in this repository is licensed under Apache License 2.0.