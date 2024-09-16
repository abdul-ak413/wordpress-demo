# Wordpress Demo

## Intro

## Installation instructions

__Documentation for reference__
- Installing docker - https://docs.docker.com/engine/install/ubuntu/
- Installing kubeadm: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
- Creating a cluster with kubeadm: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/


__Create three servers with the following settings:__

- Distribution: Ubuntu 20.04 Focal Fossa LTS

__Set appropriate hostname for each node__
E.g. k8s-control, k8s-worker1 and k8s-worker2

__On the control plane node:__

sudo hostnamectl set-hostname k8s-control

__On the first worker node:__

sudo hostnamectl set-hostname k8s-worker1

__On the second worker node__

sudo hostnamectl set-hostname k8s-worker2

__On all nodes, set up the hosts file to enable all the nodes to reach each other using these hostnames:__

sudo vi /etc/hosts

__On all nodes, add the following at the end of the file. You will need to supply the actual private IP address for each node:__

<control plane node private IP> k8s-control
<worker node 1 private IP> k8s-worker1
<worker node 2 private IP> k8s-worker2

__Log out of all three servers and log back in to see these changes take effect__

__On all nodes, set up Docker Engine and containerd. You will need to load some kernel modules and modify some system settings as part of this
process:__

cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

__sysctl params required by setup; params persist across reboots:__

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

__Apply sysctl params without reboot:__

sudo sysctl --system

__Set up the Docker Engine repository:__

__Add Docker's official GPG key:__
sudo apt-get update
sudo apt-get install ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

__Add the repository to Apt sources:__
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
  
sudo apt-get update

VERSION_STRING=5:23.0.1-1~ubuntu.20.04~focal
sudo apt-get install -y docker-ce=$VERSION_STRING docker-ce-cli=$VERSION_STRING containerd.io docker-buildx-plugin docker-compose-plugin

sudo systemctl enable docker


__Add your 'user' to the docker group:__

sudo usermod -aG docker $USER

__Log out and log back in so that your group membership is re-evaluated and that docker is up and running__

docker run hello-world

__Make sure that 'disabled_plugins' is commented out in your config.toml file:__

sudo sed -i 's/disabled_plugins/#disabled_plugins/' /etc/containerd/config.toml

__Enable and Restart containerd:__
sudo systemctl enable containerd
sudo systemctl restart containerd



__On all nodes, disable swap:__

sudo swapoff -a
sudo nano /etc/fstab

__On all nodes, install kubeadm, kubelet, and kubectl:__

curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update && sudo apt-get install -y kubelet=1.30.0-1.1 kubeadm=1.30.0-1.1 kubectl=1.30.0-1.1

sudo apt-mark hold kubelet kubeadm kubectl

sudo systemctl enable kubelet

__On the control plane node only, initialize the cluster and set up kubectl access:__

sudo kubeadm init --pod-network-cidr 10.0.0.0/16 --kubernetes-version 1.30.0

mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

__Verify the cluster is working:__

kubectl get nodes

__Install the Calico network add-on:__

kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml

__Get the join command (this command is also printed during kubeadm init; feel free to simply copy it from there):__

kubeadm token create --print-join-command

__Copy the join command from the control plane node. Run it on each worker node as root (i.e., with sudo ):__

sudo kubeadm join ...

__On the control plane node, verify all nodes in your cluster are ready. Note that it may take a few moments for all of the nodes to enter the READY state:__

kubectl get nodes
