# noctl-airship-poc
noctl PoC for Airship.
All is needed for operator - kpt

## How to run

### Preparation

If you don't have airshipctl-based gating lab the easiest way is to:

* Clone airshipctl repo
* Execute the following in the airshipctl directory:

``` sh
cd tools/gate/
00_setup.sh
10_build_gate.sh
```

If/When you have airshipctl-based gating lab configured, need to make 2 adjustments:

* make sure that MAC addresses of the air-ephemeral interfaces are the same as listed [here](https://github.com/aodinokov/noctl-airship-poc/blob/master/packages/clusters/exm01a/manifests/site/config/hosts/hosts.yaml#L39)
* Execute the following commands to make sure that VMs are accessible from docker containers:

``` sh
sudo iptables -D FORWARD -o nat_br -j REJECT --reject-with icmp-port-unreachable
sudo iptables -A FORWARD --out nat_br -s 172.17.0.0/16 -j ACCEPT
sudo iptables -A FORWARD -o nat_br -j REJECT --reject-with icmp-port-unreachable
```

### Execution

To run the PoC in the noctl-ariship-poc repo directory do the following:

``` sh
cd packages/test_packages/
./run.sh
```

The output will look like [this](https://docs.google.com/document/d/1MnCHB4lOaV9IA1x81Qe--ZmkwVIq8qvEv2B6Xmnxk5I/edit?usp=sharing).
