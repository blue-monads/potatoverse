import { usePotato } from 'potato';

export default {
  async fetch(request, env, ctx) {
    try {
      const potato = usePotato(request, env);
      
      // Basic database interaction test
      await potato.db.run_query("CREATE TABLE IF NOT EXISTS greetings (id INTEGER PRIMARY KEY, msg TEXT, ts DATETIME DEFAULT CURRENT_TIMESTAMP)");
      await potato.db.insert("greetings", { msg: "Hello from workerd executor!" });
      
      const rows = await potato.db.find_all_by_cond("greetings", {});
      const envMsg = await potato.core.get_env("MESSAGE") || "No custom message set";

      return new Response(JSON.stringify({
        status: "success",
        message: "Hello World from Potatoverse Workerd Executor!",
        env_message: envMsg,
        database_test: {
          table: "greetings",
          total_rows: rows.length,
          last_rows: rows.slice(-5)
        },
        request_info: {
          url: request.url,
          method: request.method,
          headers: Object.fromEntries(request.headers.entries())
        }
      }, null, 2), {
        headers: { 
          'Content-Type': 'application/json',
          'X-Powered-By': 'Potatoverse-Workerd'
        }
      });
    } catch (err) {
      return new Response(JSON.stringify({
        status: "error",
        error: err.message,
        stack: err.stack
      }, null, 2), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }
};
