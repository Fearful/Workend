// See https://kit.svelte.dev/docs/types#app
declare global {
  namespace App {
    interface Locals {
      user: {
        id: string;
        email: string;
        display_name: string;
        is_admin: boolean;
      } | null;
    }
    interface PageData {
      user: App.Locals['user'];
    }
  }
}

export {};
