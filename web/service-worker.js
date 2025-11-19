const CACHE_NAME = "primgo-v0.2.0";
const CACHED_FILES = [
    "/primgo/wasm_exec.js",
    "/primgo/icon512.png",
    "/primgo/favicon.png",
    "/primgo/emulator.html",
    "/primgo/service-worker.js",
    "/primgo/primgo.wasm"
];

self.addEventListener("install", (event) => {
    event.waitUntil(
        (async () => {
            const cache = await caches.open(CACHE_NAME);
            await cache.addAll(CACHED_FILES);
        })()
    );
});

self.addEventListener("activate", (event) => {
    event.waitUntil(
        caches.keys().then((keyList) => {
            return Promise.all(
                keyList.map((key) => {
                    if (key === CACHE_NAME) {
                        return;
                    }

                    return caches.delete(key);
                })
            );
        })
    );
});

self.addEventListener("fetch", (event) => {
    event.respondWith(
        (async () => {
            const cached = await caches.match(event.request);

            if (cached) {
                return cached;
            }

            const response = await fetch(event.request);

            return response;
        })()
    );
});
