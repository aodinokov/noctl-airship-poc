# noctl-airship-poc
noctl PoC for Airship


##Installation
```
git clone https://github.com/aodinokov/noctl-airship-poc.git
```

put this to ~/.profile and relogin:

```
export XDG_CONFIG_HOME=~/.kustomize
export PATH=$PATH:~/noctl-airship-poc/bin/
```

Do the following: 
```
mkdir -p $XDG_CONFIG_HOME/kustomize/plugin/someteam.example.com/v1/jinjagenerator
mkdir -p $XDG_CONFIG_HOME/kustomize/plugin/someteam.example.com/v1/jinjatransformer
cat <<EOF | tee $XDG_CONFIG_HOME/kustomize/plugin/someteam.example.com/v1/jinjagenerator/JinjaGenerator $XDG_CONFIG_HOME/kustomize/plugin/someteam.example.com/v1/jinjatransformerJinjaTransformer
#!/bin/bash
exec jinjaPlugin
EOF
chmod +x $XDG_CONFIG_HOME/kustomize/plugin/someteam.example.com/v1/jinjagenerator/JinjaGenerator $XDG_CONFIG_HOME/kustomize/plugin/someteam.example.com/v1/jinjatransformer/JinjaTransformer
```

##Work
```
cd ~
mkdir workdir
cd workdir/
kpt pkg get https://github.com/aodinokov/noctl-airship-poc.git/packages/sites/exm01a exm01a
kpt pkg sync exm01a
kustomize build --enable_alpha_plugins exm01a
```

