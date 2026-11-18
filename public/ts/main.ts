import '../css/main.css';
import '../img/favicon-32.png';
import '../img/qingyugist.svg';

import 'htmx.org';
import 'hyperscript.org';
import './code-copy';
import 'basecoat-css/basecoat';
import 'basecoat-css/dropdown-menu';
import { initGistFilters } from './gist-filter';
import { initGistLines } from './gist-lines';
import { initGistCheckboxes } from './gist-checkbox';
import { initJdenticon } from './jdenticon';
import { initOgSelects } from './og-select';
import { initPdf } from './pdf';
import { initSiteLinks } from './site-links';
import { initToasts } from './toasts';

let ipynbModule: Promise<typeof import('./ipynb')> | undefined;

const initIpynb = () => {
    if (!document.querySelector('.jupyter.notebook pre')) return;
    ipynbModule ??= import('./ipynb');
    void ipynbModule.then(({ initIpynb: renderNotebooks }) => renderNotebooks());
};

const init = () => {
    initGistFilters();
    initGistLines();
    initGistCheckboxes();
    initIpynb();
    initPdf();
    initJdenticon();
    initOgSelects();
    initSiteLinks();
    initToasts();
    syncSidebarToggle();
    syncBackToTop();
};

document.addEventListener('DOMContentLoaded', init);
// Re-init after hx-boost / htmx content swaps.
document.body.addEventListener('htmx:afterSwap', init);

// Desktop sidebar collapse toggle. The collapsed state is a class on <html>,
// which survives hx-boost body swaps, so it persists across navigation (the head
// script restores it from localStorage before paint). A delegated listener keeps
// working after the #sidebar-toggle button is swapped in on each navigation.
function syncSidebarToggle() {
    const toggle = document.getElementById('sidebar-toggle');
    if (!toggle) return;
    const collapsed = document.documentElement.classList.contains('sidebar-collapsed');
    const label = toggle.dataset[collapsed ? 'labelExpand' : 'labelCollapse'];
    if (label) {
        toggle.setAttribute('aria-label', label);
        toggle.setAttribute('title', label);
    }
}

document.addEventListener('click', (e) => {
    const toggle = (e.target as Element | null)?.closest('#sidebar-toggle');
    if (!toggle) return;
    const collapsed = document.documentElement.classList.toggle('sidebar-collapsed');
    localStorage.setItem('sidebar-collapsed', collapsed ? '1' : '0');
    syncSidebarToggle();
});

// Reveal the return-to-top control once the reader has moved past the first
// screen. The button itself is re-rendered on hx-boost swaps, so state is
// derived from scroll position rather than stored on the element.
function syncBackToTop() {
    const button = document.getElementById('back-to-top');
    if (!button) return;
    button.classList.toggle('is-visible', window.scrollY > 360);
    if (button.dataset.bound === '1') return;
    button.dataset.bound = '1';
    button.addEventListener('click', () => {
        window.scrollTo({ top: 0, behavior: 'smooth' });
    });
}

window.addEventListener('scroll', syncBackToTop, { passive: true });

// Controls marked data-auto-submit apply immediately (filters, sort order)
// instead of making the reader look for an "apply" button.
document.addEventListener('change', (e) => {
    const control = (e.target as Element | null)?.closest('[data-auto-submit]');
    if (!control) return;
    const form = control.closest('form');
    if (form) form.requestSubmit();
});

// Quick light/dark switch in the sidebar footer. This is a local override: it
// wins over the account's light/dark/auto preference and is applied before paint
// by the head script. Keep the account setting as the long-lived preference.
document.addEventListener('click', (e) => {
    const toggle = (e.target as Element | null)?.closest('#theme-toggle');
    if (!toggle) return;
    const dark = !document.documentElement.classList.contains('dark');
    document.documentElement.dataset.theme = dark ? 'dark' : 'light';
    document.documentElement.classList.toggle('dark', dark);
    localStorage.setItem('theme', dark ? 'dark' : 'light');
    window.dispatchEvent(new Event('og:themechange'));
});

// Top loading bar for boosted navigations. Large files can take a moment to
// render server-side, and htmx gives no visual feedback in the meantime, so the
// page looks frozen after a click. This animates a slim bar at the top of the
// viewport while the request is in flight.
//
// hx-boost swaps the whole <body> innerHTML on navigation, which would wipe any
// element we place in the template. We own the bar in JS instead and re-attach
// it after every swap so the reference never goes stale.
(() => {
    const bar = document.createElement('div');
    bar.id = 'page-progress';
    bar.setAttribute('aria-hidden', 'true');

    let trickle: number | undefined;
    let showTimer: number | undefined;
    let progress = 0;

    const set = (value: number) => {
        progress = value;
        bar.style.setProperty('--progress', String(value));
    };

    const attach = () => {
        // A history snapshot may have restored a stale copy of the bar (see the
        // beforeHistorySave handler below). Drop any that isn't ours.
        document.querySelectorAll('#page-progress').forEach((el) => {
            if (el !== bar) el.remove();
        });
        if (bar.parentElement !== document.body) document.body.appendChild(bar);
    };
    attach();
    document.body.addEventListener('htmx:afterSwap', attach);

    // The bar is chrome, not page content, so keep it out of the DOM snapshot
    // htmx saves for back/forward navigation — otherwise a bar frozen mid-load
    // gets restored and sticks at the top. Detaching + clearing state leaves the
    // snapshot clean; attach() re-adds a fresh bar on the way back.
    document.body.addEventListener('htmx:beforeHistorySave', () => {
        window.clearTimeout(showTimer);
        window.clearInterval(trickle);
        bar.classList.remove('is-loading');
        set(0);
        bar.remove();
    });

    const start = () => {
        window.clearTimeout(showTimer);
        window.clearInterval(trickle);
        // Delay showing so quick navigations don't flash the bar.
        showTimer = window.setTimeout(() => {
            attach();
            bar.classList.add('is-loading');
            set(0.08);
            // Creep toward (but never reach) the end while we wait.
            trickle = window.setInterval(() => {
                if (progress < 0.9) set(progress + (0.9 - progress) * 0.1);
            }, 300);
        }, 150);
    };

    const done = () => {
        window.clearTimeout(showTimer);
        window.clearInterval(trickle);
        if (!bar.classList.contains('is-loading')) return;
        // Fill to the end, then fade out. Reset the width only once it's fully
        // invisible so the bar never appears to slide backwards.
        attach();
        set(1);
        window.setTimeout(() => bar.classList.remove('is-loading'), 200);
        window.setTimeout(() => {
            bar.style.transition = 'none';
            set(0);
            void bar.offsetWidth; // flush before restoring transitions
            bar.style.transition = '';
        }, 500);
    };

    document.body.addEventListener('htmx:beforeRequest', start);
    document.body.addEventListener('htmx:afterRequest', done);
    document.body.addEventListener('htmx:historyRestore', () => {
        attach();
        done();
    });
})();

// htmx ignores error responses (4xx/5xx) by default, so a boosted navigation to
// a page that errors (e.g. a 404) would leave the user on the old page. Tell htmx
// to swap the error page in anyway so it renders like a normal navigation.
document.body.addEventListener('htmx:beforeSwap', (e) => {
    const detail = (e as CustomEvent).detail as { xhr: XMLHttpRequest; shouldSwap: boolean; isError: boolean };
    if (detail.xhr && detail.xhr.status >= 400) {
        detail.shouldSwap = true;
        detail.isError = false;
    }
});

// hx-boost stores a DOM snapshot for the back/forward cache. Components mark
// themselves initialized (basecoat: data-*-initialized, our filter:
// data-filter-ready), but the restored snapshot keeps those markers while the
// JS listeners are gone — so dropdowns etc. look dead. Strip the markers before
// the snapshot is saved, and re-initialize on restore.
const INIT_MARKERS = '[data-dropdown-menu-initialized], [data-filter-ready]';
const stripInitMarkers = () => {
    document.querySelectorAll(INIT_MARKERS).forEach((el) => {
        el.removeAttribute('data-dropdown-menu-initialized');
        el.removeAttribute('data-filter-ready');
    });
    // A saved htmx snapshot keeps the enhanced dropdown wrappers but not the JS
    // listeners inside them. Collapse them back to the native select so a
    // restore can re-enhance from scratch.
    document.querySelectorAll('.og-select').forEach((wrapper) => {
        const select = wrapper.querySelector('select');
        if (!select) return;
        select.classList.remove('og-select-native');
        delete (select as HTMLElement).dataset.ogSelectReady;
        wrapper.parentNode?.insertBefore(select, wrapper);
        wrapper.remove();
    });
    document.querySelectorAll('[data-link-editor]').forEach((editor) => {
        delete (editor as HTMLElement).dataset.linkEditorReady;
    });
};

document.body.addEventListener('htmx:beforeHistorySave', stripInitMarkers);
document.body.addEventListener('htmx:historyRestore', () => {
    stripInitMarkers();
    (window as unknown as { basecoat?: { initAll?: () => void } }).basecoat?.initAll?.();
    init();
});
