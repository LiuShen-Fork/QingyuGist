import { initOgSelects } from './og-select';

// Admin editor for the `site.links` setting. Rows are visual; a hidden input
// carries the JSON payload the server validates and stores.
const initEditor = (editor: HTMLElement) => {
    if (editor.dataset.linkEditorReady === '1') return;
    editor.dataset.linkEditorReady = '1';

    const hidden = editor.parentElement?.querySelector<HTMLInputElement>('[data-links-value]');
    const rows = editor.querySelector<HTMLElement>('[data-link-rows]');
    const template = editor.querySelector<HTMLTemplateElement>('[data-link-template]');
    const add = editor.querySelector<HTMLButtonElement>('[data-link-add]');
    if (!hidden || !rows || !template || !add) return;

    const sync = () => {
        const links = Array.from(rows.querySelectorAll<HTMLElement>('[data-link-row]'))
            .map((row) => ({
                icon: row.querySelector<HTMLSelectElement>('[data-link-icon]')?.value || 'link',
                label: row.querySelector<HTMLInputElement>('[data-link-label]')?.value.trim() || '',
                url: row.querySelector<HTMLInputElement>('[data-link-url]')?.value.trim() || '',
            }))
            .filter((link) => link.label !== '' || link.url !== '');
        hidden.value = JSON.stringify(links);
    };

    editor.addEventListener('input', sync);
    editor.addEventListener('change', sync);
    editor.closest('form')?.addEventListener('submit', sync);

    editor.addEventListener('click', (event) => {
        const target = event.target as Element | null;
        if (target?.closest('[data-link-remove]')) {
            target.closest('[data-link-row]')?.remove();
            sync();
            return;
        }
        if (!target?.closest('[data-link-add]')) return;
        const fragment = template.content.cloneNode(true) as DocumentFragment;
        const row = fragment.querySelector<HTMLElement>('[data-link-row]');
        if (!row) return;
        rows.appendChild(fragment);
        initOgSelects(row);
        row.querySelector<HTMLInputElement>('[data-link-label]')?.focus();
        sync();
    });

    sync();
};

export const initSiteLinks = (root: ParentNode = document) => {
    root.querySelectorAll<HTMLElement>('[data-link-editor]').forEach(initEditor);
};
