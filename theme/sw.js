var cacheName =  "moth:v1"
var content = [
  "index.html",
  "basic.css",
  "puzzle.js",
  "puzzle.html",
  "scoreboard.html",
  "moth.js",
  "sw.js",
  "points.json",
];

self.addEventListener("install", function(e) {
  e.waitUntil(
    caches.open(cacheName).then(function(cache) {
      return cache.addAll(content).then(
        function() {
          self.skipWaiting();
        });
    })
  );
});

/* Attempt to fetch live resources, first, then fall back to cache */
self.addEventListener('fetch', function(event) {
  let cache_used = false;

  event.respondWith(
    fetch(event.request)
     .catch(function(evt) {
      //console.log("Falling back to cache for " + event.request.url);
      cache_used = true;
      return caches.match(event.request, {ignoreSearch: true});
    }).then(function(res) {
      if (res && res.ok) {
        let res_clone = res.clone();
	if (! cache_used && event.request.method == "GET" ) {
          caches.open(cacheName).then(function(cache) {
            cache.put(event.request, res_clone);
            //console.log("Storing " + event.request.url + " in cache");
          });
	} 
        return res;
      } else {
        console.log("Failed to retrieve resource");
      }
    })
  );
});
