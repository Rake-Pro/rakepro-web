// Progressive enhancement only - the page works fully without JS.
(function () {
  "use strict";

  var reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  // Subtle pointer parallax on the .
  var = document.querySelector(".");
  if (&& !reduce) {
    window.addEventListener("pointermove", function (e) {
      var dx = (e.clientX / window.innerWidth - 0.5) * 12;
      var dy = (e.clientY / window.innerHeight - 0.5) * 12;
      .style.transform = "translate(" + dx.toFixed(1) + "px," + dy.toFixed(1) + "px)";
    }, { passive: true });
  }

  // Discord has no canonical profile URL - copy the handle to the clipboard.
  var toast = document.getElementById("toast");
  var toastTimer;
  function showToast(msg) {
    if (!toast) return;
    toast.textContent = msg;
    toast.classList.add("show");
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function () { toast.classList.remove("show"); }, 1800);
  }

  var discord = document.getElementById("discord");
  if (discord) {
    discord.addEventListener("click", function () {
      var handle = discord.getAttribute("data-handle");
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(handle).then(
          function () { showToast("Discord handle copied: " + handle); },
          function () { showToast("Discord: " + handle); }
        );
      } else {
        showToast("Discord: " + handle);
      }
    });
  }
})();
