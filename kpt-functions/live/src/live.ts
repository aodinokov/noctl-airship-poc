import {
  Configs,
  generalResult,
  //  getAnnotation,
  //  addAnnotation,
  FileFormat,
  stringify,
} from 'kpt-functions';
import { spawn } from 'child_process';

const LIVE_CMD = 'cmd';

export async function live(configs: Configs) {
  // Validate config data and read arguments.
  const args = readLiveArguments(configs);

  const cnf = new Configs();
  cnf.insert(...configs.getAll());

  try {
    const kpt = spawn('kpt', args);

    //kpt.stdin.setEncoding('utf-8');
    kpt.stdin.write(stringify(cnf, FileFormat.YAML));

    kpt.stdout.on('data', (data) => {
      console.log(`kpt stdout: ${data}`);
    });
    kpt.stderr.on('data', (data) => {
      console.log(`kpt stderr: ${data}`);
    });
    kpt.on('close', (code) => {
      if (code !== 0) {
        console.log(`kpt process exited with code ${code}`);
	}
	//resolve();
    });
  }catch (err) {
    configs.addResults(generalResult(err, 'error'));
  }
}

function readLiveArguments(configs: Configs) {
  const args: string[] = [];
  const configMap = configs.getFunctionConfigMap();
  if (!configMap) {
    return args;
  }
  configMap.forEach((value: string, key: string) => {
    if (key === LIVE_CMD) {
      args.push('--' + key);
    //} else if (key === VERBOSE || key === IGNORE_MAC) {
    //  args.push('--' + key);
    } else {
      args.push('--' + key);
      args.push(value);
    }
  });
  return args;
}

live.usage = `
Sops function (see https://github.com/mozilla/sops).
So far supports only decrypt operation:
runs sops -d for all documents that have field 'sops:' and put the decrypted result back.

Can be configured using a ConfigMap with the following flags:
ignore-mac: true [Optional: default empty] Ignore Message Authentication Code during decryption.
verbose: true [Optional: default empty]    Enable sops verbose logging output.
keyservice value [Optional: default empty] Specify the key services to use in addition to the local one.
                                           Can be specified more than once.
                                           Syntax: protocol://address. Example: tcp://myserver.com:5000
override-detached-annotations: [Optional:
default see detachedAnnotations var]       The list of annotations that didn't present when the document
                                           was encrypted, but added by different tools later. The function
                                           will detach them before decryption and added unchanged
                                           after successfull decryption. This allows sops to check the
                                           consistency of the decrypted document.

For more details see 'sops --help'.

Example:

To decrypt the documents use:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/sops
    config.kubernetes.io/local-config: "true"
data:
  verbose: true
`;
