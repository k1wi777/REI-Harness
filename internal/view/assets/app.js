// app.js — JavaScript propio mínimo del visor (sin dependencias de terceros).
// Se encarga de: (1) recarga en vivo por SSE y (2) el botón Aprobar.
(function () {
  "use strict";

  function setNotice(text) {
    var el = document.getElementById("notice");
    if (el) {
      el.textContent = text || "";
    }
  }

  function refresh() {
    fetch("/content", { headers: { "Accept": "text/html" } })
      .then(function (r) { return r.text(); })
      .then(function (html) {
        var el = document.getElementById("content");
        if (el) {
          el.innerHTML = html;
        }
      })
      .catch(function () { /* recarga best-effort */ });
  }

  function approve() {
    var btn = document.getElementById("approve");
    if (btn) {
      btn.disabled = true;
    }
    setNotice("Aprobando…");
    fetch("/approve", { method: "POST" })
      .then(function (r) {
        return r.json().then(function (data) { return { ok: r.ok, data: data }; });
      })
      .then(function (res) {
        if (res.ok) {
          setNotice(res.data.message || "Aprobado.");
        } else {
          setNotice("No se pudo aprobar: " + (res.data.error || "error"));
          var b = document.getElementById("approve");
          if (b) {
            b.disabled = false;
          }
        }
      })
      .catch(function () {
        setNotice("Error de red al aprobar.");
        var b = document.getElementById("approve");
        if (b) {
          b.disabled = false;
        }
      });
  }

  // Delegación de eventos: el botón puede reaparecer tras una recarga.
  document.addEventListener("click", function (ev) {
    var t = ev.target;
    if (t && t.id === "approve") {
      approve();
    }
  });

  // SSE: el servidor emite `data: reload` al cambiar los documentos.
  try {
    var es = new EventSource("/events");
    es.onmessage = function (ev) {
      if (ev.data === "reload") {
        refresh();
      }
    };
  } catch (e) {
    // Sin SSE la página sigue siendo útil; solo no se recarga sola.
  }
})();
