import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
const home = process.env.USERPROFILE
const proj = 'C:/Users/gbloc/Projects/AgentOrange'
const files = {
  userSettings: join(home, '.claude', 'settings.json'),
  claudeJson: join(home, '.claude.json'),
  projSettings: join(proj, '.claude', 'settings.json'),
  localSettings: join(proj, '.claude', 'settings.local.json'),
  mcpJson: join(proj, '.mcp.json'),
}
const out = { parse: {}, keys: {} }
const load = (p) => { try { return JSON.parse(readFileSync(p, 'utf8')) } catch (e) { return { __err: String(e).slice(0, 120) } } }
for (const [k, p] of Object.entries(files)) {
  if (!existsSync(p)) { out.parse[k] = 'absent'; continue }
  const j = load(p)
  out.parse[k] = j.__err ? 'PARSE FAIL: ' + j.__err : 'ok'
  if (j.__err) continue
  if (k === 'userSettings') {
    out.keys.channel = j.autoUpdatesChannel || 'unset(latest)'
    out.keys.userDefaultMode = (j.permissions && j.permissions.defaultMode) || 'unset'
    out.keys.userEnabledPlugins = j.enabledPlugins ? Object.keys(j.enabledPlugins) : []
    out.keys.userHooks = j.hooks ? Object.keys(j.hooks) : []
    out.keys.disableAutoMode = (j.permissions && j.permissions.disableAutoMode) || j.disableAutoMode || 'unset'
    out.keys.userAllowCount = (j.permissions && j.permissions.allow && j.permissions.allow.length) || 0
  }
  if (k === 'claudeJson') {
    out.keys.installMethod = j.installMethod || 'unset'
    out.keys.numStartups = j.numStartups
    out.keys.autoUpdates = j.autoUpdates !== undefined ? j.autoUpdates : 'unset'
    out.keys.skillUsage = j.skillUsage ? Object.entries(j.skillUsage).map(([n, v]) => [n, v.usageCount, v.lastUsedAt]) : []
    out.keys.pluginUsage = j.pluginUsage ? Object.entries(j.pluginUsage).map(([n, v]) => [n, v.usageCount, v.lastUsedAt]) : []
    out.keys.userMcp = j.mcpServers ? Object.keys(j.mcpServers) : []
    const projEntry = j.projects && j.projects[proj.replace(/\//g, '\\')] || j.projects && j.projects[proj]
    out.keys.localMcp = projEntry && projEntry.mcpServers ? Object.keys(projEntry.mcpServers) : []
    out.keys.disabledMcp = projEntry && projEntry.disabledMcpServers ? projEntry.disabledMcpServers : []
    out.keys.projectKeys = j.projects ? Object.keys(j.projects).length : 0
  }
  if (k === 'localSettings') {
    out.keys.localDefaultMode = (j.permissions && j.permissions.defaultMode) || 'unset'
    out.keys.localAllow = (j.permissions && j.permissions.allow) || []
    out.keys.localEnabledPlugins = j.enabledPlugins ? Object.entries(j.enabledPlugins) : []
    out.keys.localHooks = j.hooks ? Object.keys(j.hooks) : []
  }
  if (k === 'mcpJson') out.keys.projMcp = j.mcpServers ? Object.keys(j.mcpServers) : []
}
console.log(JSON.stringify(out, null, 1))
