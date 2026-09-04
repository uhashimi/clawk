package sandbox

// herdrIntegrationsScript runs at guest boot (as the sandbox user, after
// the agent state dirs are mounted) and installs herdr's built-in
// agent-state integrations for every coding agent the image actually
// ships. The integrations write a hook into each agent's state dir that
// reports its live session state to a running herdr (herdr sets
// HERDR_ENV=1 plus a socket and pane id when an agent runs inside one of
// its panes; outside a herdr pane the hooks are no-ops).
//
// Why at boot rather than at image build time: the agent state dirs are
// host-mounted at boot, so anything a build wrote under $HOME would be
// shadowed by the mount; and "which agents are installed" is only
// knowable in the guest — the host-side state dirs exist for every
// registered runner whether or not the image ships the binary.
//
// Why every boot rather than once: herdr's install is idempotent — it
// "ensures" each hook and re-merges claude's settings.json entry while
// preserving the other keys (clawk's own forced settings survive) — so
// re-running refreshes hooks when the user updates herdr and re-adds one
// the user deleted. The script is best-effort end to end: no herdr in
// the image (or a failed install) just skips, and a command failure
// never holds boot (see guestcfg.Command).
//
// The agent list is the intersection of herdr's built-in integrations
// (claude, codex, pi, omp, opencode, …) and clawk's runners. prime-agent
// has no herdr integration upstream. pi and omp additionally need their
// extensions dir to exist — herdr validates the agent's layout and
// refuses an agent that hasn't run yet — hence the mkdir -p steps.
//
// Scope: only the OCI/vz manifest (OCIGuestManifest) carries this command.
// That is the path that mounts the per-sandbox agent state dirs; the
// firecracker provider mounts none (no per-runner persistence there), so
// its hookless rootfs just means `herdr integration install` is run by
// hand if wanted.
const herdrIntegrationsScript = `
command -v herdr >/dev/null 2>&1 || exit 0
command -v pi >/dev/null 2>&1 && mkdir -p "$HOME/.pi/agent/extensions"
command -v omp >/dev/null 2>&1 && mkdir -p "$HOME/.omp/agent/extensions"
for agent in claude codex pi omp opencode; do
  command -v "$agent" >/dev/null 2>&1 || continue
  herdr integration install "$agent" >/dev/null 2>&1 || true
done
exit 0
`
