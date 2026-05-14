import { e as escape_html, c as ensure_array_like, a as attr } from "../../../chunks/renderer.js";
import { P as PageHeader } from "../../../chunks/PageHeader.js";
import { P as Panel } from "../../../chunks/Panel.js";
import { B as Badge } from "../../../chunks/Badge.js";
import { F as FlashMessage } from "../../../chunks/FlashMessage.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    function providerLabel(kind) {
      switch (kind) {
        case "github":
          return "GitHub";
        case "gitlab":
          return "GitLab";
        case "gitea":
          return "Gitea";
        default:
          return kind;
      }
    }
    function providerVariant(kind) {
      switch (kind) {
        case "github":
          return "info";
        case "gitlab":
          return "warning";
        case "gitea":
          return "success";
        default:
          return "muted";
      }
    }
    const WEBHOOK_EVENTS = [
      "run_started",
      "run_completed",
      "run_failed",
      "project_created",
      "member_added"
    ];
    PageHeader($$renderer2, { title: "Settings" });
    $$renderer2.push(`<!----> <p class="nav-link svelte-1i19ct2"><a href="/settings/notifications">→ Notification targets</a></p> `);
    if (data.flash.connected) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Connected to ${escape_html(data.flash.connected)}.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.flash.error) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Connection failed: ${escape_html(data.flash.error)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    Panel($$renderer2, {
      title: "Account",
      children: ($$renderer3) => {
        $$renderer3.push(`<div class="row svelte-1i19ct2"><span class="label svelte-1i19ct2">Display name</span><span class="value svelte-1i19ct2">${escape_html(data.user?.display_name)}</span><span></span></div> <div class="row svelte-1i19ct2"><span class="label svelte-1i19ct2">Email</span><span class="value svelte-1i19ct2">${escape_html(data.user?.email)}</span><span></span></div>`);
      }
    });
    $$renderer2.push(`<!----> <div class="panels-grid two-col svelte-1i19ct2">`);
    Panel($$renderer2, {
      title: "Git provider connections",
      children: ($$renderer3) => {
        if (!data.providersConfigured || data.connections.length === 0) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="hint svelte-1i19ct2">No OAuth providers configured on this Workend instance. Set <code>WORKEND_GITHUB_CLIENT_ID</code> / <code>WORKEND_GITLAB_CLIENT_ID</code> / <code>WORKEND_GITEA_CLIENT_ID</code> (with matching <code>_CLIENT_SECRET</code> and <code>WORKEND_TOKEN_KEY</code>) in compose to enable.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array = ensure_array_like(data.connections);
          for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
            let c = each_array[$$index];
            $$renderer3.push(`<div class="row svelte-1i19ct2">`);
            Badge($$renderer3, {
              variant: providerVariant(c.provider),
              size: "sm",
              children: ($$renderer4) => {
                $$renderer4.push(`<!---->${escape_html(providerLabel(c.provider))}`);
              }
            });
            $$renderer3.push(`<!----> <div><div class="value svelte-1i19ct2">${escape_html(c.instance_host)}</div> `);
            if (c.connected) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<p class="hint svelte-1i19ct2">@${escape_html(c.handle)} · scopes: ${escape_html(c.scopes || "—")}</p>`);
            } else {
              $$renderer3.push("<!--[-1-->");
              $$renderer3.push(`<p class="hint svelte-1i19ct2">not connected</p>`);
            }
            $$renderer3.push(`<!--]--></div> `);
            if (c.connected) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<form method="POST" action="?/disconnect" class="inline-form"><input type="hidden" name="id"${attr("value", c.connection_id || "")}/> <button type="submit" class="ghost">Disconnect</button></form>`);
            } else {
              $$renderer3.push("<!--[-1-->");
              $$renderer3.push(`<button type="button">Connect</button>`);
            }
            $$renderer3.push(`<!--]--></div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Panel($$renderer2, {
      title: "HTTPS access tokens",
      children: ($$renderer3) => {
        $$renderer3.push(`<p class="hint svelte-1i19ct2">Used for cloning private repos via <code>https://host/user/repo.git</code> URLs. Paste a personal access token from your git host. The token is
      encrypted at rest and never shown again after saving.</p> `);
        if (data.patCredentials.length === 0) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="hint svelte-1i19ct2">No access tokens yet.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array_1 = ensure_array_like(data.patCredentials);
          for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
            let p = each_array_1[$$index_1];
            $$renderer3.push(`<div class="row row-stretch svelte-1i19ct2"><div><div class="value svelte-1i19ct2">${escape_html(p.host)}</div> <p class="hint svelte-1i19ct2">${escape_html(p.label)} · added ${escape_html(new Date(p.created_at).toLocaleDateString())}</p></div> <form method="POST" action="?/deletePAT" class="inline-form"><input type="hidden" name="id"${attr("value", p.id)}/> <button type="submit" class="ghost">Delete</button></form></div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]--> <form method="POST" action="?/addPAT" style="margin-top: var(--space-4);"><div class="field"><label for="pat-host">Host</label> <input id="pat-host" name="host" type="text" required="" placeholder="e.g. github.com or gitlab.example.com"${attr("value", form?.patHost || "")}/></div> <div class="field"><label for="pat-label">Label</label> <input id="pat-label" name="label" type="text" required="" placeholder="e.g. personal-pat (read:repo)"${attr("value", form?.patLabel || "")}/></div> <div class="field"><label for="pat-token">Token</label> <input id="pat-token" name="token" type="password" required="" autocomplete="off" placeholder="ghp_… / glpat-… / etc."/></div> `);
        if (form?.patError) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="form-error svelte-1i19ct2">${escape_html(form.patError)}</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> `);
        if (form?.patAdded) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="form-success svelte-1i19ct2">Access token saved.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> <button type="submit">Save access token</button></form>`);
      }
    });
    $$renderer2.push(`<!----></div> `);
    Panel($$renderer2, {
      title: "SSH keys",
      children: ($$renderer3) => {
        $$renderer3.push(`<p class="hint svelte-1i19ct2">Used for cloning repos via <code>git@host:user/repo.git</code> URLs. Paste an
    unencrypted PEM private key and add the matching public key to your git host.</p> `);
        if (data.sshKeys.length === 0) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="hint svelte-1i19ct2">No SSH keys yet.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array_2 = ensure_array_like(data.sshKeys);
          for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
            let k = each_array_2[$$index_2];
            $$renderer3.push(`<div class="row row-three svelte-1i19ct2"><div><div class="value svelte-1i19ct2">${escape_html(k.name)}</div> <p class="hint ssh-key svelte-1i19ct2">${escape_html(k.public_key)}</p></div> <button type="button" class="ghost">Copy public key</button> <form method="POST" action="?/deleteSSHKey" class="inline-form"><input type="hidden" name="id"${attr("value", k.id)}/> <button type="submit" class="ghost">Delete</button></form></div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]--> <form method="POST" action="?/addSSHKey" style="margin-top: var(--space-4);"><div class="field"><label for="ssh-name">Key name</label> <input id="ssh-name" name="name" type="text" required="" placeholder="e.g. laptop-ed25519"${attr("value", form?.sshName || "")}/></div> <div class="field"><label for="ssh-private">Private key (PEM)</label> <textarea id="ssh-private" name="private_key" rows="6" required="" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----
..."></textarea></div> `);
        if (form?.sshError) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="form-error svelte-1i19ct2">${escape_html(form.sshError)}</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> `);
        if (form?.sshAdded) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="form-success svelte-1i19ct2">SSH key added.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> <button type="submit">Add SSH key</button></form>`);
      }
    });
    $$renderer2.push(`<!----> `);
    Panel($$renderer2, {
      title: "Outbound webhooks",
      children: ($$renderer3) => {
        $$renderer3.push(`<p class="hint svelte-1i19ct2">Receive HTTP POST callbacks when events happen in your workspaces.
    Optionally set a secret to verify webhook signatures.</p> `);
        if (data.webhooks.length === 0) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="hint svelte-1i19ct2">No webhooks configured yet.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array_3 = ensure_array_like(data.webhooks);
          for (let $$index_4 = 0, $$length = each_array_3.length; $$index_4 < $$length; $$index_4++) {
            let wh = each_array_3[$$index_4];
            $$renderer3.push(`<div class="row row-stretch svelte-1i19ct2"><div><div class="webhook-url svelte-1i19ct2">${escape_html(wh.url)}</div> <div class="event-badges svelte-1i19ct2" style="margin-top: var(--space-1);"><!--[-->`);
            const each_array_4 = ensure_array_like(wh.events);
            for (let $$index_3 = 0, $$length2 = each_array_4.length; $$index_3 < $$length2; $$index_3++) {
              let ev = each_array_4[$$index_3];
              Badge($$renderer3, {
                variant: "muted",
                size: "sm",
                children: ($$renderer4) => {
                  $$renderer4.push(`<!---->${escape_html(ev)}`);
                }
              });
            }
            $$renderer3.push(`<!--]--></div> <p class="hint svelte-1i19ct2">added ${escape_html(new Date(wh.created_at).toLocaleDateString())}${escape_html(wh.secret_hash ? " · signed" : "")}</p></div> <div class="webhook-actions svelte-1i19ct2"><form method="POST" action="?/testWebhook" class="inline-form"><input type="hidden" name="id"${attr("value", wh.id)}/> <button type="submit" class="ghost">Test</button></form> <form method="POST" action="?/deleteWebhook" class="inline-form"><input type="hidden" name="id"${attr("value", wh.id)}/> <button type="submit" class="ghost">Delete</button></form></div></div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]--> `);
        if (form?.webhookTested) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="form-success svelte-1i19ct2">Test event sent.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> `);
        if (form?.webhookDeleted) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="form-success svelte-1i19ct2">Webhook deleted.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> <form method="POST" action="?/addWebhook" style="margin-top: var(--space-4);"><div class="field"><label for="wh-url">Payload URL</label> <input id="wh-url" name="url" type="url" required="" placeholder="https://example.com/webhook"${attr("value", form?.webhookUrl || "")}/></div> <fieldset class="field" style="border: none; padding: 0; margin: 0 0 1rem 0;"><legend style="display: block; margin-bottom: 0.25rem; font-size: 0.875rem; color: var(--text-muted);">Events</legend> <div class="checkbox-group svelte-1i19ct2"><!--[-->`);
        const each_array_5 = ensure_array_like(WEBHOOK_EVENTS);
        for (let $$index_5 = 0, $$length = each_array_5.length; $$index_5 < $$length; $$index_5++) {
          let ev = each_array_5[$$index_5];
          $$renderer3.push(`<label class="svelte-1i19ct2"><input type="checkbox" name="events"${attr("value", ev)} class="svelte-1i19ct2"/> ${escape_html(ev)}</label>`);
        }
        $$renderer3.push(`<!--]--></div></fieldset> <div class="field"><label for="wh-secret">Secret (optional)</label> <input id="wh-secret" name="secret" type="password" autocomplete="off" placeholder="Used for HMAC signature verification"/></div> `);
        if (form?.webhookError) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="form-error svelte-1i19ct2">${escape_html(form.webhookError)}</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> `);
        if (form?.webhookAdded) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="form-success svelte-1i19ct2">Webhook created.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> <button type="submit">Add webhook</button></form>`);
      }
    });
    $$renderer2.push(`<!----> `);
    Panel($$renderer2, {
      title: "Push notifications",
      children: ($$renderer3) => {
        {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="push-info svelte-1i19ct2">Push notifications are not supported in this browser. Use a modern browser with
      service worker support to enable push notifications.</div>`);
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}
export {
  _page as default
};
