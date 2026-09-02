(function () {
  var slice = function (x) { return Array.prototype.slice.call(x); };
  var tabs = slice(document.querySelectorAll('nav.tabs button'));
  var docs = slice(document.querySelectorAll('main article'));
  var filter = document.getElementById('filter');
  var toc = document.getElementById('toc');
  var tocbtn = document.getElementById('tocbtn');
  var backdrop = document.getElementById('tocbackdrop');

  function active() { return docs.filter(function (d) { return !d.hidden; })[0]; }

  // ---- table of contents: a sidebar on wide screens, a drawer behind the Contents button
  // on narrow ones — same list, built from the ACTIVE document's h2 spine ----
  var tocLinks = [];
  function closeToc() {
    document.body.classList.remove('toc-open');
    if (backdrop) { backdrop.hidden = true; }
    if (tocbtn) { tocbtn.setAttribute('aria-expanded', 'false'); }
  }
  function buildToc(d) {
    if (!toc || !d) { return; }
    var hs = slice(d.querySelectorAll('h2'));
    toc.innerHTML = ''; tocLinks = [];
    var thin = hs.length < 2;
    toc.hidden = thin;
    if (tocbtn) { tocbtn.hidden = thin; }
    if (thin) { closeToc(); return; }
    var title = document.createElement('p');
    title.className = 'toc-title'; title.textContent = 'Contents';
    toc.appendChild(title);
    var ol = document.createElement('ol');
    hs.forEach(function (h, i) {
      if (!h.id) { h.id = 's-' + (d.dataset.doc || 'doc').replace(/[^a-z0-9]/gi, '') + '-' + i; }
      var li = document.createElement('li'); var a = document.createElement('a');
      a.href = '#' + h.id; a.textContent = h.textContent;
      a.addEventListener('click', closeToc);
      li.appendChild(a); ol.appendChild(li); tocLinks.push({ a: a, h: h });
    });
    toc.appendChild(ol);
    spy();
  }
  function spy() {
    if (!tocLinks.length) { return; }
    var current = tocLinks[0];
    tocLinks.forEach(function (t) { if (t.h.getBoundingClientRect().top <= 90) { current = t; } });
    tocLinks.forEach(function (t) { t.a.className = t === current ? 'active' : ''; });
  }
  var ticking = false;
  window.addEventListener('scroll', function () {
    if (ticking) { return; }
    ticking = true;
    requestAnimationFrame(function () { spy(); ticking = false; });
  }, { passive: true });
  if (tocbtn) {
    tocbtn.addEventListener('click', function () {
      var open = document.body.classList.toggle('toc-open');
      if (backdrop) { backdrop.hidden = !open; }
      tocbtn.setAttribute('aria-expanded', String(open));
    });
  }
  if (backdrop) { backdrop.addEventListener('click', closeToc); }
  document.addEventListener('keydown', function (e) { if (e.key === 'Escape') { closeToc(); closePop(); } });

  function show(file, frag) {
    tabs.forEach(function (t) { t.setAttribute('aria-selected', String(t.dataset.doc === file)); });
    docs.forEach(function (d) { d.hidden = d.dataset.doc !== file; });
    closePop();
    closeToc();
    if (frag) {
      var el = document.getElementById(frag);
      if (el) { el.scrollIntoView({ block: 'start' }); }
    } else {
      window.scrollTo(0, 0);
    }
    if (history.replaceState) { history.replaceState(null, '', '#' + file + (frag ? '/' + frag : '')); }
    apply();
    buildToc(active());
    mermaidize(active());
  }

  tabs.forEach(function (t) { t.addEventListener('click', function () { show(t.dataset.doc, ''); }); });
  tabs.forEach(function (t, i) {
    t.addEventListener('keydown', function (e) {
      var j = e.key === 'ArrowRight' ? i + 1 : e.key === 'ArrowLeft' ? i - 1 : -1;
      if (j < 0 || j >= tabs.length) { return; }
      tabs[j].focus(); show(tabs[j].dataset.doc, '');
    });
  });

  // A cross-document id link switches the tab, THEN jumps. This is the whole reason the site
  // exists: in the markdown set the same link is a file the reader has to open and search.
  document.addEventListener('click', function (e) {
    var a = e.target.closest ? e.target.closest('a.idref') : null;
    if (!a) { return; }
    var doc = a.dataset.doc;
    var frag = a.getAttribute('href').slice(1);
    if (doc) { e.preventDefault(); show(doc, frag); }
  });

  // ---- citations: a tap on the numbered marker opens the References entry IN PLACE ----
  var pop = null;
  function closePop() { if (pop && pop.parentNode) { pop.parentNode.removeChild(pop); } pop = null; }
  document.addEventListener('click', function (e) {
    var a = e.target.closest ? e.target.closest('a.cite') : null;
    if (!a) { return; }
    var art = active();
    if (!art) { return; }
    var id = a.getAttribute('href').slice(1);
    var li = art.querySelector('[id="' + id + '"]');
    if (!li) { return; }
    e.preventDefault();
    var again = pop && pop.dataset.fn === id;
    closePop();
    if (again) { return; }
    pop = document.createElement('div');
    pop.className = 'fnpop'; pop.dataset.fn = id;
    pop.innerHTML = li.innerHTML;
    var host = e.target.closest('p, li, td, blockquote') || a.parentNode;
    host.insertAdjacentElement('afterend', pop);
  });

  // Filtering hides SECTIONS, not lines: a transcript filtered to matching sentences is a
  // different document. A section is a heading and everything under it up to the next one of
  // the same or higher rank.
  function apply() {
    var q = (filter.value || '').trim().toLowerCase();
    var act = active();
    if (!act) { return; }
    var kids = slice(act.children);
    var group = [], groupText = '', groups = [];
    kids.forEach(function (el) {
      if (/^H[1-3]$/.test(el.tagName) && group.length) {
        groups.push({ els: group, text: groupText });
        group = []; groupText = '';
      }
      group.push(el);
      groupText += ' ' + (el.textContent || '').toLowerCase();
    });
    if (group.length) { groups.push({ els: group, text: groupText }); }
    groups.forEach(function (g) {
      var hit = !q || g.text.indexOf(q) >= 0;
      g.els.forEach(function (el) { el.classList.toggle('hidden-by-filter', !hit); });
    });
  }
  filter.addEventListener('input', apply);

  // ---- mermaid: OPTIONAL enhancement over a page that is already complete without it ----
  // The import is dynamic and pinned; offline or blocked, the fenced source stays exactly
  // where it was with a note saying what it is. new Function keeps the syntax out of old
  // parsers so the rest of this script survives them.
  var dynImport = null;
  try { dynImport = new Function('u', 'return import(u)'); } catch (err) { dynImport = null; }
  var merLib = null;
  function mermaidize(d) {
    if (!d) { return; }
    var codes = slice(d.querySelectorAll('pre > code.language-mermaid')).filter(function (c) {
      return !c.parentNode.dataset.mmDone;
    });
    if (!codes.length) { return; }
    if (!dynImport) { codes.forEach(function (c) { note(c.parentNode); }); return; }
    merLib = merLib || dynImport('MERMAID_CDN_URL').then(function (m) {
      var mm = m.default;
      mm.initialize({
        startOnLoad: false, securityLevel: 'strict',
        theme: window.matchMedia && matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'neutral'
      });
      return mm;
    });
    merLib.then(function (mm) {
      codes.forEach(function (c, i) {
        var pre = c.parentNode;
        if (pre.dataset.mmDone) { return; }
        pre.dataset.mmDone = '1';
        mm.render('mm-' + Date.now() + '-' + i, c.textContent).then(function (r) {
          var fig = document.createElement('figure');
          fig.className = 'mermaid'; fig.innerHTML = r.svg;
          var det = document.createElement('details');
          det.innerHTML = '<summary>diagram source</summary>';
          pre.parentNode.insertBefore(fig, pre);
          det.appendChild(pre); fig.appendChild(det);
        }).catch(function () { note(pre); });
      });
    }).catch(function () {
      merLib = null;
      codes.forEach(function (c) { note(c.parentNode); });
    });
  }
  function note(pre) {
    if (pre.dataset.mmNote) { return; }
    pre.dataset.mmNote = '1';
    var p = document.createElement('p');
    p.className = 'mermaid-note';
    p.textContent = 'diagram source — rendering it needs network access this copy did not have';
    pre.parentNode.insertBefore(p, pre);
  }

  var hash = (location.hash || '').slice(1);
  var parts = hash.split('/');
  if (hash && docs.some(function (d) { return d.dataset.doc === parts[0]; })) {
    show(parts[0], parts[1] || '');
  } else {
    buildToc(active());
    mermaidize(active());
  }
})();
