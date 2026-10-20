(function (root, factory) {
  if (typeof module === 'object' && module.exports) module.exports = factory();
  else root.HexisDataCache = factory();
})(typeof globalThis !== 'undefined' ? globalThis : this, function () {
  const scopes = new Map();
  function create(scope, request, options) {
    options = options || {};
    const now = options.now || Date.now;
    const ttl = options.ttl ?? 5 * 60 * 1000;
    const versionTTL = options.versionTTL ?? 60 * 1000;
    if (!scopes.has(scope)) scopes.set(scope, new Map());
    const entries = scopes.get(scope);
    function cached(path) {
      const previous = entries.get(path);
      if (previous && (previous.pending || previous.expires > now())) return previous.promise;
      const entry = { pending: true, expires: 0 };
      entry.promise = Promise.resolve().then(() => request(path)).then(value => {
        entry.pending = false;
        entry.expires = now() + (path === '/api/v1/patches/latest' ? versionTTL : ttl);
        return value;
      }, error => {
        if (entries.get(path) === entry) entries.delete(path);
        throw error;
      });
      entries.set(path, entry);
      return entry.promise;
    }
    cached.invalidate = () => entries.clear();
    cached.isStale = () => {
      const version = entries.get('/api/v1/patches/latest');
      return !version || (!version.pending && version.expires <= now());
    };
    return cached;
  }
  return { create };
});
