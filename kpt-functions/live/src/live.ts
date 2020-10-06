import {
  Configs,
  generalResult,
  getLabel,
  FileFormat,
  stringify,
} from 'kpt-functions';
import { ChildProcess, spawn } from 'child_process';
import { ConfigMap, isConfigMap } from './gen/io.k8s.api.core.v1';

const INVENTORY_LABEL = 'cli-utils.sigs.k8s.io/inventory-id';

export const CMD = 'cmd';
const CMD_ALLOWED = new Map<string, string>([
  ['source', 'fn'],
  ['sink', 'fn'],
  ['apply', 'live'],
  ['destroy', 'live'],
  ['status', 'live'],
  ['diff', 'live'],
  ['preview', 'live'],
]);
export const PATH = 'path';
export const INVENTORY_OBJECT_NAME = 'inventoryObjectName';
export const INVENTORY_OBJECT_NAMESPACE = 'inventoryObjectNamespace';
export const INVENTORY_ID = 'inventoryId';
const INVENTORY_SETTINGS = new Map<string, string>();

function trimNewLine(s: string): string {
  return s.toString().replace(/(\r\n|\n|\r)$/gm, '');
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

function createInventory(): ConfigMap | undefined {
  const name = INVENTORY_SETTINGS.get(INVENTORY_OBJECT_NAME);
  const namespace = INVENTORY_SETTINGS.get(INVENTORY_OBJECT_NAMESPACE);
  const id = INVENTORY_SETTINGS.get(INVENTORY_ID);

  if (name === undefined || namespace === undefined || id === undefined) {
    return undefined;
  }

  return new ConfigMap({
    metadata: {
      name: '{name}',
      namespace: '{namespace}',
      labels: { [INVENTORY_LABEL]: id },
    },
  });
}

function getLiveConfigs(configs: Configs): Configs {
  const cfgs = new Configs();
  cfgs.insert(...configs.getAll());

  let foundInventory = false;
  cfgs.get(isConfigMap).forEach((n) => {
    if (getLabel(n, INVENTORY_LABEL) !== undefined) {
      foundInventory = true;
      console.log(`I: found inventory`);
    }
  });

  if (!foundInventory) {
    const inv = createInventory();
    if (inv !== undefined) {
      console.log(`I: added inventory`);
      cfgs.insert(inv);
    } else {
      console.log(
        `W: wasn't able to create inventory(some params are missing?)`
      );
    }
  }

  return cfgs;
}

export async function live(configs: Configs) {
  // Validate config data and read arguments.
  const args = readLiveArguments(configs);
  const cfgs = getLiveConfigs(configs);

  try {
    const kpt = spawn('kpt', args);

    kpt.stdin.write(stringify(cfgs, FileFormat.YAML));
    kpt.stdin.end();

    kpt.stdout.on('data', (data) => {
      console.log(`I: ${trimNewLine(data)}`);
    });
    kpt.stderr.on('data', (data) => {
      console.log(`E: ${trimNewLine(data)}`);
    });
    await onExit(kpt);
  } catch (err) {
    console.log(`kpt run finished with error: ${err}`);
    configs.addResults(generalResult(err, 'error'));
  }
}

function readLiveArguments(configs: Configs) {
  let cmd: string | undefined = undefined;
  let path: string | undefined = undefined;
  const args: string[] = [];
  const result: string[] = [];

  const configMap = configs.getFunctionConfigMap();
  if (!configMap) {
    return result;
  }

  configMap.forEach((value: string, key: string) => {
    if (key === CMD) {
      cmd = value;
    } else if (key === PATH) {
      path = value;
    } else if (
      key === INVENTORY_OBJECT_NAME ||
      key === INVENTORY_OBJECT_NAMESPACE ||
      key === INVENTORY_ID
    ) {
      INVENTORY_SETTINGS.set(key, value);
    } else {
      args.push('--' + key);
      args.push(value);
    }
  });

  // building up resulting array
  if (cmd === undefined) {
    return result;
  }
  const prefix = CMD_ALLOWED.get(cmd);
  if (prefix === undefined) {
    return result;
  }
  result.push(prefix);
  result.push(cmd);

  if (path !== undefined) {
    result.push(path);
  }
  result.push(...args);

  return result;
}

live.usage = `
Live function runs kpt live commands.
Additionally it allows to run source and sink commands.

Can be configured using a ConfigMap with the following flags:
cmd:  [Mandatory: apply|destroy|status|diff|preview|sink|source]
      Defines the command that will be executed.
path: [Optional] defined the path that will be used to execute the command.
inventoryObjectName,
inventoryObjectNamespace,
inventoryId: [Optional] if set and there is no inventory ConfigMap found in
      the input, the function will add a ConfigMap with
      the corresponding name, namespace and with label
      cli-utils.sigs.k8s.io/inventory-id = inventoryId value. This action
      is similar to kpt live init (see https://googlecontainertools.github.io/kpt/reference/live/init/).
all other params: [Optional] along with values will be added as arguments in the following
      format --<key> <value>.
Example:
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: quay.io/aodinokov/live:v0.0.1 
    config.kubernetes.io/local-config: "true"
data:
  cmd: apply
  inventoryObjectName: inventory-18149771
  inventoryObjectNamespace: default
  inventoryId: a6ec3136-30a8-4bd5-a2d7-ccde1433f114
  reconcile-timeout: 15m
`;
