export interface Config {
  LocalPort: number
  RemotePort: number
  SSHHost: string
  SSHUser: string
  SSHPort: number
  LocalHost: string
  RemoteCmd: string
  DisconnectCmd: string
}

export interface TunnelInfo {
  id: string
  localHost: string
  localPort: number
  remotePort: number
  sshUser: string
  sshHost: string
  sshPort: number
  remoteCmd: string
  disconnectCmd: string
}
