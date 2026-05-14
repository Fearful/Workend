import './state.svelte-DaEZ9E9Z.js';
import './root-BPP4VMgG.js';
import { w as writable } from './index-rPPP9l7g.js';

function create_updated_store() {
  const { set, subscribe } = writable(false);
  {
    return {
      subscribe,
      // eslint-disable-next-line @typescript-eslint/require-await
      check: async () => false
    };
  }
}
const stores = {
  updated: /* @__PURE__ */ create_updated_store()
};
function goto(url, opts = {}) {
  {
    throw new Error("Cannot call goto(...) on the server");
  }
}
({
  check: stores.updated.check
});

export { goto as g };
//# sourceMappingURL=client-Bp0-LT_C.js.map
