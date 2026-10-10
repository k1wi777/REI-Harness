// app.js — JavaScript propio mínimo del visor (sin dependencias de terceros).
// Se encarga de: (1) recarga en vivo por SSE, (2) el selector de documento del
// índice (solo se muestra el documento elegido en la columna de lectura) y
// (3) el botón Aprobar.
(function () {
  "use strict";

  var STORE_KEY = "rei.view.doc";
  var IDX_KEY = "rei.view.index";
  var smallViewport = window.matchMedia("(max-width: 1080px)");
  var indexCollapsedState = false;

  function indexStored() {
    try {
      var v = localStorage.getItem(IDX_KEY);
      if (v !== null) {
        return v === "1";
      }
    } catch (e) {
      /* sin almacenamiento */
    }
    // Por defecto: colapsado en móvil (menú desplegable), abierto en escritorio.
    return smallViewport.matches;
  }

  // setIndex muestra u oculta el índice; con `remember` persiste la elección.
  function setIndex(collapsed, remember) {
    indexCollapsedState = collapsed;
    if (remember) {
      try {
        localStorage.setItem(IDX_KEY, collapsed ? "1" : "0");
      } catch (e) {
        /* sin almacenamiento */
      }
    }
    document.body.classList.toggle("index-collapsed", collapsed);
    var btn = document.getElementById("index-toggle");
    if (btn) {
      btn.setAttribute("aria-expanded", collapsed ? "false" : "true");
    }
  }

  function setNotice(text) {
    var el = document.getElementById("notice");
    if (el) {
      el.textContent = text || "";
    }
  }

  function storedDoc() {
    try {
      return localStorage.getItem(STORE_KEY);
    } catch (e) {
      return null;
    }
  }

  function storeDoc(id) {
    try {
      localStorage.setItem(STORE_KEY, id);
    } catch (e) {
      /* almacenamiento no disponible; la selección no persiste */
    }
  }

  // applySelection muestra el documento activo y marca el índice.
  function applySelection() {
    var reader = document.querySelector(".reader");
    if (!reader) {
      return;
    }
    var bodies = reader.querySelectorAll(".doc-body");
    if (!bodies.length) {
      return;
    }
    var wanted = storedDoc();
    var found = false;
    for (var i = 0; i < bodies.length; i++) {
      if (bodies[i].getAttribute("data-doc") === wanted) {
        found = true;
        break;
      }
    }
    if (!found) {
      wanted = bodies[0].getAttribute("data-doc");
    }
    for (var j = 0; j < bodies.length; j++) {
      bodies[j].classList.toggle("is-active", bodies[j].getAttribute("data-doc") === wanted);
    }
    var items = document.querySelectorAll(".doc");
    for (var k = 0; k < items.length; k++) {
      var link = items[k].querySelector("a.nm");
      items[k].classList.toggle("active", !!link && link.getAttribute("data-doc") === wanted);
    }
  }

  function selectDoc(id) {
    storeDoc(id);
    if (window.history && history.replaceState) {
      history.replaceState(null, "", "#doc-" + id);
    }
    applySelection();
    if (smallViewport.matches && !indexCollapsedState) {
      setIndex(true, true);
    }
    var reader = document.querySelector(".reader");
    if (reader) {
      reader.scrollTop = 0;
    }
  }

  function refresh() {
    fetch("/content", { headers: { "Accept": "text/html" } })
      .then(function (r) { return r.text(); })
      .then(function (html) {
        var el = document.getElementById("content");
        if (el) {
          el.innerHTML = html;
          applySelection();
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

  // Delegación de eventos: el botón y los enlaces del índice reaparecen tras
  // una recarga, así que no se enlazan directamente.
  document.addEventListener("click", function (ev) {
    var t = ev.target;
    if (!t || !t.closest) {
      return;
    }
    var link = t.closest(".doc a.nm");
    if (link) {
      var id = link.getAttribute("data-doc");
      if (id) {
        ev.preventDefault();
        selectDoc(id);
      }
      return;
    }
    if (t.id === "approve" || t.closest("#approve")) {
      approve();
    } else if (t.id === "index-toggle" || t.closest("#index-toggle")) {
      setIndex(!indexCollapsedState, true);
    }
  });

  // Activa el modo "una columna, un documento" y aplica la selección inicial.
  document.body.classList.add("js-viewer");
  setIndex(indexStored(), false);
  applySelection();

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
