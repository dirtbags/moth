function pwa_init() {
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register("./sw.js").then(function(reg) {
    }).catch(function(err) {
      console.warn("Error while registering service worker", err);
    });
  } else {
    console.log("Service workers not supported. Some offline functionality may not work");
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", pwa_init);
} else {
  pwa_init();
}
