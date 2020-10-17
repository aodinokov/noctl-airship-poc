import { Configs, generalResult, isKubernetesObject } from 'kpt-functions';
import rw from 'rw';
import { ChildProcess, spawn } from 'child_process';
import { safeLoadAll } from 'js-yaml';

const CONFIG_PATH = '/tmp/clusterctl.yaml';

const INLINE_CONFIG_ARG_NAME = 'inlineConfig';
const CMD_ARG_NAME = 'cmd';
const STDOUT_TO_PIPELINE_ARG_NAME = 'stdoutToPipeline';

let stdoutToPipeline = false;

function trimNewLine(s: string): string {
  return s.toString().replace(/(\r\n|\n|\r)$/gm, '');
}

async function writeFile(path: string, data: string): Promise<void> {
  return new Promise((resolve, reject) => {
    rw.writeFile(path, data, 'utf8', (err: object) => {
      if (err) return reject(err);
      resolve();
    });
  });
}

function onExit(childProcess: ChildProcess): Promise<void> {
  return new Promise((resolve, reject) => {
    childProcess.once('exit', (code: number, signal: string) => {
      if (code === 0) {
        resolve(undefined);
      } else {
        reject(new Error('Exit with error code: ' + code));
      }
    });
    childProcess.once('error', (err: Error) => {
      reject(err);
    });
  });
}

export async function clusterctl(configs: Configs) {
  const args = readArguments(configs);

  try {
    let prcsStdout = '';
    const prcs = spawn('clusterctl', args);

    prcs.stdin.end();

    prcs.stdout.on('data', (data) => {
      if (stdoutToPipeline) {
        prcsStdout = prcsStdout + data;
      } else {
        console.log(`I: ${trimNewLine(data)}`);
      }
    });
    prcs.stderr.on('data', (data) => {
      console.log(`E: ${trimNewLine(data)}`);
    });

    await onExit(prcs);

    if (stdoutToPipeline) {
      let objects = safeLoadAll(prcsStdout);
      objects = objects.filter((o) => isKubernetesObject(o));
      configs.insert(...objects);
    }
  } catch (err) {
    console.log(`clusterctl run finished with error: ${err}`);
    configs.addResults(generalResult(err, 'error'));
  }
}

function readArguments(configs: Configs) {
  let args: string[] = [];
  const configMap = configs.getFunctionConfigMap();
  if (!configMap) {
    return args;
  }
  let cmd: string = "";
  let config: string = "";
  const cmdParams: string[] = [];
  configMap.forEach((value: string, key: string) => {
    if (key === CMD_ARG_NAME) {
      cmd = value;
    } else if (key === INLINE_CONFIG_ARG_NAME) {
      config = value;
    } else if (key === STDOUT_TO_PIPELINE_ARG_NAME) {
      if (value === 'true') {
        stdoutToPipeline = true;
      }
    } else {
      if (key.startsWith("--")) {
        cmdParams.push(key+"="+value);
      } else {
        cmdParams.push(key);
        cmdParams.push(value);
      }
    }
  });

  if (cmd === "") {
    return args;
  }
  args = args.concat(cmd)
  args = args.concat(cmdParams)

  if (config !== "") {
    writeFile(CONFIG_PATH, config);   
    args = args.concat("--config="+CONFIG_PATH); 
  }
  console.log(`${args}`);
  return args;
}

clusterctl.usage = `
Execute clusterctl bin with parameters specified. 
Configured using a ConfigMap with a key for {${CMD_ARG_NAME}}.
Works with arbitrary clusterctl commands like init and flags like --infrastructure:
${CMD_ARG_NAME}: command, can contain several words, e.g. "config cluster".
--infrastructure: [Optional] sets the type of infrastructure.
...
Example:
To init a Openstack provider:
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/clusterctl
    config.kubernetes.io/local-config: "true"
data:
  ${CMD_ARG_NAME}: "init"
  --infrastructure: "openstack"
`;
