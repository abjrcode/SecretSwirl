import { LogDebug, LogError, LogInfo, LogWarning } from '../wailsjs/runtime'

const consoleLog = console.log
const consoleWarn = console.warn
const consoleError = console.error
const consoleDebug = console.debug

console.log = function (msg: string) {
  consoleLog(msg)
  LogInfo(msg)
}

console.warn = function (msg: string) {
  consoleWarn(msg)
  LogWarning(msg)
}

console.error = function (msg: string) {
  consoleError(msg)
  LogError(msg)
}

console.debug = function (msg: string) {
  consoleDebug(msg)
  LogDebug(msg)
}