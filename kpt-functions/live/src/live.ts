import {
  Configs,
  generalResult,
  //  getAnnotation,
  //  addAnnotation,
  FileFormat,
  stringify,
} from 'kpt-functions';
import { ChildProcess, spawn } from 'child_process';

export const LIVE_CMD = 'cmd';

function trimNewLine(s: string): string {
  return s.toString().replace(/(\r\n|\n|\r)$/gm, "");
}

function onExit(childProcess: ChildProcess): Promise<void> {
  return new Promise((resolve, reject) => {
    childProcess.once('exit', (code: number, signal: string) => {
      if (code === 0) {
        resolve(undefined);
      } else {
        reject(new Error('Exit with error code: '+code));
      }
    });
    childProcess.once('error', (err: Error) => {
      reject(err);
    });
  });
}

export async function live(configs: Configs) {
  // Validate config data and read arguments.
  const args = readLiveArguments(configs);

  const cnf = new Configs();
  cnf.insert(...configs.getAll());

  try {
    const kpt = spawn('kpt', args);

    kpt.stdin.write(stringify(cnf, FileFormat.YAML));
    kpt.stdin.end();

    kpt.stdout.on('data', (data) => {
      console.log(`I: ${trimNewLine(data)}`);
    });
    kpt.stderr.on('data', (data) => {
      console.log(`E: ${trimNewLine(data)}`);
    });
    await onExit(kpt);
  }catch (err) {
    console.log(`kpt run finished with error: ${err}`);
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
      args.push('live');
      args.push(value);
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
`;
