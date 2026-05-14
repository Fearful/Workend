<script lang="ts">
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';

  let { data, form } = $props();

  // ---------- Constants ----------
  const RESOURCES = ['project', 'run', 'pipeline', 'secret', 'member', 'role', 'workspace'] as const;
  const ACTIONS = ['read', 'create', 'update', 'delete'] as const;

  // ---------- Roles State ----------
  let showCreateRole = $state(false);
  let editingRoleId = $state<string | null>(null);
  let deleteRoleId = $state<string | null>(null);

  interface Role {
    id: string;
    workspace_id: string;
    name: string;
    description: string;
    is_system: boolean;
    permissions: { resource: string; action: string }[];
    created_at: string;
  }

  let editingRole = $derived(
    editingRoleId ? data.roles.find((r: Role) => r.id === editingRoleId) : null
  );
  let deletingRole = $derived(
    deleteRoleId ? data.roles.find((r: Role) => r.id === deleteRoleId) : null
  );

  function hasPermission(permissions: { resource: string; action: string }[], resource: string, action: string): boolean {
    return permissions.some(p => p.resource === resource && p.action === action);
  }

  // ---------- Signing Keys State ----------
  let showPrivateKey = $state(false);
  let confirmRevokeKeyId = $state<string | null>(null);

  let revokingKey = $derived(
    confirmRevokeKeyId ? data.signingKeys.find((k: { id: string }) => k.id === confirmRevokeKeyId) : null
  );

  // ---------- Effects ----------
  $effect(() => {
    if (form?.roleCreated) showCreateRole = false;
    if (form?.roleUpdated) editingRoleId = null;
    if (form?.roleDeleted) deleteRoleId = null;
    if (form?.keyGenerated) showPrivateKey = true;
    if (form?.keyRevoked) confirmRevokeKeyId = null;
  });
</script>

<style>
  h1 {
    font-size: var(--fs-2xl);
    margin: 0 0 0.25rem 0;
    font-weight: var(--fw-semibold);
    letter-spacing: -0.02em;
    line-height: var(--lh-tight);
  }
  .role-tag {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    margin-left: var(--space-2);
    font-weight: var(--fw-regular);
  }
  .header-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
    flex-wrap: wrap;
  }
  .ws-tabs {
    border-bottom: 1px solid var(--border);
    display: flex;
    gap: 0;
    margin-bottom: var(--space-5);
  }
  .ws-tab {
    padding: 0.625rem 0.875rem;
    color: var(--text-dim);
    text-decoration: none;
    font-size: var(--fs-md);
    font-weight: var(--fw-medium);
    border-bottom: 2px solid transparent;
    margin-bottom: -1px;
  }
  .ws-tab:hover { color: var(--text); text-decoration: none; }
  .ws-tab.active { color: var(--text); border-bottom-color: var(--accent); }

  .section { margin-top: var(--space-6); }

  /* ---- Roles List ---- */
  .role-card {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-4) var(--space-5);
    margin-bottom: var(--space-3);
    transition: border-color 120ms ease;
  }
  .role-card:hover { border-color: var(--border-strong); }
  .role-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-2);
  }
  .role-name {
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .role-desc {
    color: var(--text-muted);
    font-size: var(--fs-sm);
    margin-bottom: var(--space-3);
  }
  .role-actions {
    display: flex;
    gap: var(--space-1);
    flex-shrink: 0;
  }

  /* ---- Permission Grid ---- */
  .perm-grid {
    display: grid;
    grid-template-columns: 120px repeat(4, 1fr);
    gap: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
    font-size: var(--fs-xs);
  }
  .perm-header {
    background: var(--bg-page);
    font-weight: var(--fw-semibold);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-dim);
    padding: var(--space-2) var(--space-2);
    border-bottom: 1px solid var(--border);
  }
  .perm-resource {
    background: var(--bg-page);
    font-weight: var(--fw-medium);
    padding: var(--space-2) var(--space-3);
    border-bottom: 1px solid var(--border);
    border-right: 1px solid var(--border);
    color: var(--text);
  }
  .perm-cell {
    padding: var(--space-2);
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .perm-cell:last-child { border-right: none; }
  .perm-grid > :nth-last-child(-n+5) { border-bottom: none; }

  .perm-check {
    color: var(--status-success-fg);
    font-weight: 600;
  }
  .perm-empty { color: var(--text-dim); opacity: 0.3; }

  /* ---- Checkbox Grid (for forms) ---- */
  .perm-grid-form {
    display: grid;
    grid-template-columns: 120px repeat(4, 1fr);
    gap: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
    font-size: var(--fs-xs);
  }
  .perm-grid-form .perm-cell { cursor: pointer; }
  .perm-grid-form .perm-cell:hover { background: var(--bg-hover); }
  .perm-grid-form input[type="checkbox"] {
    accent-color: var(--accent);
    width: 14px;
    height: 14px;
  }

  /* ---- Create Role Form ---- */
  .create-form {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-5) var(--space-6);
    margin-bottom: var(--space-4);
  }
  .create-form h3 {
    margin: 0 0 var(--space-3);
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
  }
  .form-row {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
  }
  .form-row .field { margin: 0; }
  .form-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-4);
  }

  /* ---- Members Role Assignment ---- */
  .member-list {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
  }
  .member-assign-row {
    display: grid;
    grid-template-columns: 1fr 1fr auto;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-3) var(--space-5);
    border-bottom: 1px solid var(--border);
    font-size: var(--fs-sm);
  }
  .member-assign-row:last-child { border-bottom: none; }
  .member-assign-name { font-weight: var(--fw-medium); }
  .member-assign-email {
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
  }
  .role-assign-form {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }
  .role-assign-select {
    padding: 0.375rem 0.625rem;
    background: var(--bg-panel);
    color: var(--text);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    font: inherit;
    font-size: var(--fs-xs);
  }

  /* ---- Signing Keys ---- */
  .key-list {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
  }
  .key-row {
    display: grid;
    grid-template-columns: 1fr auto auto auto;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-3) var(--space-5);
    border-bottom: 1px solid var(--border);
    font-size: var(--fs-sm);
  }
  .key-row:last-child { border-bottom: none; }
  .key-hash {
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    color: var(--text);
  }
  .key-meta {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    margin-top: 2px;
  }
  .key-revoked {
    opacity: 0.5;
  }

  .private-key-display {
    background: var(--bg-page);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 240px;
    overflow-y: auto;
    margin: var(--space-3) 0;
  }
  .key-warning {
    color: var(--status-danger-fg);
    font-size: var(--fs-sm);
    font-weight: var(--fw-medium);
    margin-bottom: var(--space-2);
  }

  @media (max-width: 768px) {
    .perm-grid, .perm-grid-form { font-size: 0.625rem; grid-template-columns: 80px repeat(4, 1fr); }
    .form-row { grid-template-columns: 1fr; }
    .member-assign-row { grid-template-columns: 1fr auto; }
    .member-assign-email { display: none; }
    .key-row { grid-template-columns: 1fr auto; }
    .key-row > :nth-child(2),
    .key-row > :nth-child(3) { display: none; }
  }
</style>

<Breadcrumb segments={[
  { label: 'workspaces', href: '/' },
  { label: data.workspace.name, href: `/workspaces/${data.workspace.id}` },
  { label: 'roles' }
]} />

<div class="header-row">
  <h1>{data.workspace.name} <span class="role-tag">{data.workspace.my_role}</span></h1>
</div>

<nav class="ws-tabs" aria-label="Workspace sections">
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}`}>Overview</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/dashboard`}>Dashboard</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/activity`}>Activity</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/secrets`}>Secrets</a>
  <a class="ws-tab active" href={`/workspaces/${data.workspace.id}/roles`}>Roles</a>
</nav>

{#if form?.roleError}
  <FlashMessage type="error">{form.roleError}</FlashMessage>
{/if}
{#if form?.assignError}
  <FlashMessage type="error">{form.assignError}</FlashMessage>
{/if}
{#if form?.keyError}
  <FlashMessage type="error">{form.keyError}</FlashMessage>
{/if}
{#if form?.roleCreated}
  <FlashMessage type="success">Role created.</FlashMessage>
{/if}
{#if form?.roleUpdated}
  <FlashMessage type="success">Role updated.</FlashMessage>
{/if}
{#if form?.roleDeleted}
  <FlashMessage type="success">Role deleted.</FlashMessage>
{/if}
{#if form?.roleAssigned}
  <FlashMessage type="success">Role assigned.</FlashMessage>
{/if}
{#if form?.keyRevoked}
  <FlashMessage type="success">Signing key revoked.</FlashMessage>
{/if}

{#if data.rolesError}
  <FlashMessage type="error">{data.rolesError}</FlashMessage>
{/if}

<!-- ===== ROLES SECTION ===== -->
<SectionHeader title="Roles">
  {#snippet actions()}
    <button onclick={() => { showCreateRole = !showCreateRole; }}>
      {showCreateRole ? 'Cancel' : 'Create role'}
    </button>
  {/snippet}
</SectionHeader>

{#if showCreateRole}
  <div class="create-form">
    <h3>New role</h3>
    <form method="POST" action="?/createRole">
      <div class="form-row">
        <div class="field">
          <label for="role-name">Name</label>
          <input id="role-name" name="name" type="text" required placeholder="deployer" value={form?.roleName || ''} />
        </div>
        <div class="field">
          <label for="role-desc">Description</label>
          <input id="role-desc" name="description" type="text" placeholder="Can deploy and manage runs" />
        </div>
      </div>
      <span id="perm-label-create" style="display: block; margin-bottom: var(--space-2); font-size: var(--fs-sm); font-weight: var(--fw-medium);">Permissions</span>
      <div class="perm-grid-form" role="group" aria-labelledby="perm-label-create">
        <div class="perm-header"></div>
        {#each ACTIONS as action}
          <div class="perm-header" style="text-align: center;">{action}</div>
        {/each}
        {#each RESOURCES as resource}
          <div class="perm-resource">{resource}</div>
          {#each ACTIONS as action}
            <label class="perm-cell">
              <input type="checkbox" name="permissions" value="{resource}:{action}" aria-label="{resource} {action}" />
            </label>
          {/each}
        {/each}
      </div>
      <div class="form-footer">
        <button type="button" class="ghost" onclick={() => { showCreateRole = false; }}>Cancel</button>
        <button type="submit">Create role</button>
      </div>
    </form>
  </div>
{/if}

{#if data.roles.length === 0 && !showCreateRole}
  <EmptyState icon="◎" message="No roles defined. Create custom roles to control access." />
{:else}
  {#each data.roles as role (role.id)}
    <div class="role-card">
      <div class="role-head">
        <span class="role-name">
          {role.name}
          {#if role.is_system}<Badge variant="muted" size="sm">System</Badge>{/if}
        </span>
        {#if !role.is_system}
          <div class="role-actions">
            <button class="ghost" onclick={() => { editingRoleId = role.id; }}>Edit</button>
            <button class="ghost" onclick={() => { deleteRoleId = role.id; }}>Delete</button>
          </div>
        {/if}
      </div>
      {#if role.description}
        <div class="role-desc">{role.description}</div>
      {/if}
      <div class="perm-grid">
        <div class="perm-header"></div>
        {#each ACTIONS as action}
          <div class="perm-header" style="text-align: center;">{action}</div>
        {/each}
        {#each RESOURCES as resource}
          <div class="perm-resource">{resource}</div>
          {#each ACTIONS as action}
            <div class="perm-cell">
              {#if hasPermission(role.permissions, resource, action)}
                <span class="perm-check">&#10003;</span>
              {:else}
                <span class="perm-empty">&mdash;</span>
              {/if}
            </div>
          {/each}
        {/each}
      </div>
    </div>
  {/each}
{/if}

<!-- ===== MEMBER ROLE ASSIGNMENT ===== -->
<div class="section">
  <SectionHeader title="Member roles" />
  {#if data.members.length === 0}
    <EmptyState message="No members to assign roles to." />
  {:else}
    <div class="member-list">
      {#each data.members as member (member.user_id)}
        <div class="member-assign-row">
          <span class="member-assign-name">{member.display_name}</span>
          <span class="member-assign-email">{member.email}</span>
          <form method="POST" action="?/assignRole" class="role-assign-form">
            <input type="hidden" name="user_id" value={member.user_id} />
            <select name="role_id" class="role-assign-select">
              {#each data.roles as role (role.id)}
                <option value={role.id}>{role.name}</option>
              {/each}
            </select>
            <button type="submit" class="ghost">Assign</button>
          </form>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- ===== SIGNING KEYS SECTION ===== -->
<div class="section">
  <SectionHeader title="Signing keys">
    {#snippet actions()}
      <form method="POST" action="?/generateKey" style="display: inline;">
        <button type="submit">Generate key</button>
      </form>
    {/snippet}
  </SectionHeader>

  {#if data.keysError}
    <FlashMessage type="error">{data.keysError}</FlashMessage>
  {/if}

  {#if data.signingKeys.length === 0}
    <EmptyState icon="⚿" message="No signing keys. Generate a key pair to sign deployments." />
  {:else}
    <div class="key-list">
      {#each data.signingKeys as key (key.id)}
        <div class="key-row" class:key-revoked={key.revoked_at !== null}>
          <div>
            <div class="key-hash">{key.key_hash}</div>
            <div class="key-meta">
              by {key.created_by} &middot; <TimeAgo value={key.created_at} />
            </div>
          </div>
          {#if key.revoked_at}
            <Badge variant="danger" size="sm">Revoked</Badge>
          {:else}
            <Badge variant="success" size="sm">Active</Badge>
          {/if}
          <span style="font-family: var(--font-mono); font-size: var(--fs-xs); color: var(--text-dim);">
            {key.public_key.slice(0, 24)}...
          </span>
          {#if !key.revoked_at}
            <button class="ghost" onclick={() => { confirmRevokeKeyId = key.id; }}>Revoke</button>
          {:else}
            <span></span>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- ===== MODALS ===== -->

<!-- Edit Role Modal -->
<Modal open={editingRoleId !== null && editingRole !== undefined} title="Edit role" width={560} onClose={() => { editingRoleId = null; }}>
  {#if editingRole}
    <form method="POST" action="?/updateRole">
      <input type="hidden" name="role_id" value={editingRole.id} />
      <div class="field" style="margin-bottom: var(--space-3);">
        <label for="edit-role-name">Name</label>
        <input id="edit-role-name" name="name" type="text" value={editingRole.name} />
      </div>
      <div class="field" style="margin-bottom: var(--space-3);">
        <label for="edit-role-desc">Description</label>
        <input id="edit-role-desc" name="description" type="text" value={editingRole.description} />
      </div>
      <span id="perm-label-edit" style="display: block; margin-bottom: var(--space-2); font-size: var(--fs-sm); font-weight: var(--fw-medium);">Permissions</span>
      <div class="perm-grid-form" role="group" aria-labelledby="perm-label-edit">
        <div class="perm-header"></div>
        {#each ACTIONS as action}
          <div class="perm-header" style="text-align: center;">{action}</div>
        {/each}
        {#each RESOURCES as resource}
          <div class="perm-resource">{resource}</div>
          {#each ACTIONS as action}
            <label class="perm-cell">
              <input type="checkbox" name="permissions" value="{resource}:{action}"
                     checked={hasPermission(editingRole.permissions, resource, action)} aria-label="{resource} {action}" />
            </label>
          {/each}
        {/each}
      </div>
      <div class="form-footer">
        <button type="button" class="ghost" onclick={() => { editingRoleId = null; }}>Cancel</button>
        <button type="submit">Save changes</button>
      </div>
    </form>
  {/if}
</Modal>

<!-- Delete Role Confirmation Modal -->
<Modal open={deleteRoleId !== null && deletingRole !== undefined} title="Delete role" onClose={() => { deleteRoleId = null; }}>
  {#if deletingRole}
    <p>Delete the role <strong>{deletingRole.name}</strong>? Members with this role will lose these permissions.</p>
    <form method="POST" action="?/deleteRole">
      <input type="hidden" name="role_id" value={deletingRole.id} />
      <div style="display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-4);">
        <button type="button" class="ghost" onclick={() => { deleteRoleId = null; }}>Cancel</button>
        <button type="submit" style="background: var(--danger); border-color: var(--danger);">Delete</button>
      </div>
    </form>
  {/if}
</Modal>

<!-- Revoke Key Confirmation Modal -->
<Modal open={confirmRevokeKeyId !== null && revokingKey !== undefined} title="Revoke signing key" onClose={() => { confirmRevokeKeyId = null; }}>
  {#if revokingKey}
    <p>Revoke key <strong style="font-family: var(--font-mono);">{revokingKey.key_hash}</strong>? Deployments signed with this key will no longer be verified.</p>
    <form method="POST" action="?/revokeKey">
      <input type="hidden" name="key_id" value={revokingKey.id} />
      <div style="display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-4);">
        <button type="button" class="ghost" onclick={() => { confirmRevokeKeyId = null; }}>Cancel</button>
        <button type="submit" style="background: var(--danger); border-color: var(--danger);">Revoke</button>
      </div>
    </form>
  {/if}
</Modal>

<!-- Private Key Display Modal -->
<Modal open={showPrivateKey && !!form?.privateKey} title="Private key generated" width={560} onClose={() => { showPrivateKey = false; }}>
  <div class="key-warning">Copy this private key now. It will not be shown again.</div>
  <div class="private-key-display">{form?.privateKey}</div>
  <div style="display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-3);">
    <button onclick={() => {
      if (form?.privateKey) navigator.clipboard.writeText(form.privateKey);
    }}>Copy to clipboard</button>
    <button class="ghost" onclick={() => { showPrivateKey = false; }}>Done</button>
  </div>
</Modal>
