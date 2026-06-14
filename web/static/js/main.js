// Progressive enhancement only - the page is fully functional without JS.
(function () {
  "use strict";

  // Subtle parallax on the hero halo, tied to pointer position. Disabled when
  // the user prefers reduced motion.
  var reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  var halo = document.querySelector(".halo");

  if (halo && !reduce) {
    window.addEventListener("pointermove", function (e) {
      var dx = (e.clientX / window.innerWidth - 0.5) * 14;
      var dy = (e.clientY / window.innerHeight - 0.5) * 14;
      halo.style.transform = "translate(" + dx.toFixed(1) + "px," + dy.toFixed(1) + "px)";
    }, { passive: true });
  }

  // Reveal feature cards as they scroll into view.
  var cards = document.querySelectorAll(".card");
  if ("IntersectionObserver" in window && cards.length) {
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          entry.target.style.opacity = "1";
          entry.target.style.transform = "translateY(0)";
          io.unobserve(entry.target);
        }
      });
    }, { threshold: 0.2 });

    cards.forEach(function (card) {
      card.style.opacity = "0";
      card.style.transform = "translateY(18px)";
      card.style.transition = "opacity 0.6s ease, transform 0.6s ease";
      io.observe(card);
    });
  }
})();
