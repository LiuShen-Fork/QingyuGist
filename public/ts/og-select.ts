// Progressive enhancement for native <select> elements. The original select
// stays in the form (so submission, validation, hyperscript and existing
// `onchange` handlers keep working) and we render a styled listbox on top of
// it. This gives every dropdown in the app the same shadcn-style surface
// without pulling a component framework into the server-rendered stack.

type OptionEntry = {
    value: string;
    label: string;
    disabled: boolean;
    element: HTMLButtonElement;
};

let activeClose: (() => void) | null = null;

const cloneWidthClasses = (select: HTMLSelectElement, wrapper: HTMLElement) => {
    for (const className of Array.from(select.classList)) {
        if (
            className === 'w-full' ||
            className === 'shrink-0' ||
            className === 'text-xs' ||
            className.startsWith('max-w-') ||
            className.startsWith('sm:max-w-')
        ) {
            wrapper.classList.add(className);
        }
    }
};

// Hide the native box as an inline style as well as a class. The class lives in
// the compiled stylesheet, and a restored htmx snapshot (or a stylesheet still
// in flight) can leave a select that carries `og-select-native` — or a stale
// `ogSelectReady` marker without its wrapper — visible right next to the styled
// trigger. Belt and braces: the element that must never be seen is hidden by
// the DOM itself.
const hideNative = (select: HTMLSelectElement) => {
    select.classList.add('og-select-native');
    select.style.setProperty('display', 'none');
};

const showNative = (select: HTMLSelectElement) => {
    select.classList.remove('og-select-native');
    select.style.removeProperty('display');
};

const enhanceSelect = (select: HTMLSelectElement) => {
    // A select marked ready that no longer sits in its wrapper is a leftover:
    // the wrapper went away (snapshot restore, DOM patched by htmx) and with it
    // every listener, while the marker survived. Reset it and enhance afresh.
    if (select.dataset.ogSelectReady === '1' && !select.parentElement?.classList.contains('og-select')) {
        delete select.dataset.ogSelectReady;
        showNative(select);
    }

    if (select.dataset.ogSelectReady === '1') {
        hideNative(select);
        return;
    }

    if (select.multiple || select.size > 1) return;
    if (select.closest('.og-select')) {
        hideNative(select);
        return;
    }

    select.dataset.ogSelectReady = '1';

    const wrapper = document.createElement('div');
    wrapper.className = 'og-select';
    cloneWidthClasses(select, wrapper);

    const trigger = document.createElement('button');
    trigger.type = 'button';
    trigger.className = 'og-select-trigger';
    trigger.setAttribute('aria-haspopup', 'listbox');
    trigger.setAttribute('aria-expanded', 'false');
    const value = document.createElement('span');
    value.className = 'og-select-value';
    const valueId = `og-select-value-${Math.random().toString(36).slice(2, 9)}`;
    value.id = valueId;

    const chevron = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    chevron.setAttribute('viewBox', '0 0 24 24');
    chevron.setAttribute('fill', 'none');
    chevron.setAttribute('stroke', 'currentColor');
    chevron.setAttribute('stroke-width', '2');
    chevron.setAttribute('stroke-linecap', 'round');
    chevron.setAttribute('stroke-linejoin', 'round');
    chevron.setAttribute('aria-hidden', 'true');
    chevron.classList.add('og-select-chevron');
    const chevronPath = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    chevronPath.setAttribute('d', 'm6 9 6 6 6-6');
    chevron.appendChild(chevronPath);

    trigger.append(value, chevron);

    // Borrow the <label for="..."> text as the button's accessible name.
    if (select.id) {
        const label = document.querySelector<HTMLLabelElement>(`label[for="${CSS.escape(select.id)}"]`);
        if (label) {
            if (!label.id) label.id = `og-label-${select.id}`;
            trigger.setAttribute('aria-labelledby', `${label.id} ${valueId}`);
        }
    }
    if (!trigger.hasAttribute('aria-labelledby')) {
        trigger.setAttribute('aria-label', select.getAttribute('aria-label') || select.name || 'Select');
    }

    const menu = document.createElement('div');
    menu.className = 'og-select-menu';
    menu.id = `${valueId}-menu`;
    menu.setAttribute('role', 'listbox');
    menu.hidden = true;
    trigger.setAttribute('aria-controls', menu.id);

    const entries: OptionEntry[] = [];
    const addOption = (option: HTMLOptionElement) => {
        if (option.hidden) return;
        const item = document.createElement('button');
        item.type = 'button';
        item.className = 'og-select-option';
        item.setAttribute('role', 'option');
        item.dataset.value = option.value;
        item.disabled = option.disabled;

        const check = document.createElement('span');
        check.className = 'og-select-check';
        check.innerHTML =
            '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m5 12 4 4L19 6"/></svg>';
        const text = document.createElement('span');
        text.className = 'og-select-option-label';
        text.textContent = option.textContent?.trim() || option.value;
        item.append(check, text);
        menu.appendChild(item);
        entries.push({ value: option.value, label: text.textContent || '', disabled: option.disabled, element: item });
    };

    for (const child of Array.from(select.children)) {
        if (child instanceof HTMLOptGroupElement) {
            const group = document.createElement('div');
            group.className = 'og-select-group';
            group.textContent = child.label;
            menu.appendChild(group);
            for (const option of Array.from(child.querySelectorAll('option'))) addOption(option);
        } else if (child instanceof HTMLOptionElement) {
            addOption(child);
        }
    }

    const sync = () => {
        const selected = select.selectedOptions[0];
        value.textContent = selected ? (selected.textContent || '').trim() : '';
        trigger.disabled = select.disabled;
        wrapper.classList.toggle('is-disabled', select.disabled);
        for (const entry of entries) {
            const active = selected ? entry.value === selected.value : false;
            entry.element.setAttribute('aria-selected', active ? 'true' : 'false');
        }
    };

    const close = (focusTrigger = false) => {
        if (menu.hidden) return;
        menu.hidden = true;
        wrapper.classList.remove('is-open');
        trigger.setAttribute('aria-expanded', 'false');
        document.removeEventListener('pointerdown', onPointerDown, true);
        window.removeEventListener('resize', onViewportChange);
        window.removeEventListener('scroll', onViewportChange, true);
        if (activeClose === close) activeClose = null;
        if (focusTrigger) trigger.focus();
    };

    const position = () => {
        const rect = trigger.getBoundingClientRect();
        const margin = 8;
        const menuWidth = Math.min(Math.max(rect.width, 128), window.innerWidth - margin * 2);
        menu.style.width = `${menuWidth}px`;
        menu.style.left = `${Math.max(margin, Math.min(rect.left, window.innerWidth - menuWidth - margin))}px`;
        menu.style.top = '0px';
        menu.style.maxHeight = 'none';

        const menuHeight = menu.offsetHeight;
        const below = window.innerHeight - rect.bottom - margin;
        const above = rect.top - margin;
        const openUp = below < menuHeight && above > below;
        if (openUp) {
            menu.style.top = `${Math.max(margin, rect.top - menuHeight - 4)}px`;
            menu.style.maxHeight = `${Math.max(120, above - 4)}px`;
        } else {
            menu.style.top = `${rect.bottom + 4}px`;
            menu.style.maxHeight = `${Math.max(120, below - 4)}px`;
        }
    };

    const onPointerDown = (event: PointerEvent) => {
        if (!wrapper.contains(event.target as Node)) close();
    };
    const onViewportChange = (event: Event) => {
        // Scrolling inside the menu itself is not a reason to dismiss it.
        if (event.type === 'scroll' && event.target instanceof Node && menu.contains(event.target)) return;
        close();
    };

    const open = (focusLast = false) => {
        if (select.disabled || menu.hidden === false) return;
        if (activeClose && activeClose !== close) activeClose();
        activeClose = close;
        menu.hidden = false;
        wrapper.classList.add('is-open');
        trigger.setAttribute('aria-expanded', 'true');
        sync();
        position();

        const selected = entries.find((entry) => entry.value === select.value && !entry.disabled);
        const target = focusLast ? [...entries].reverse().find((entry) => !entry.disabled) : selected;
        (target?.element ?? entries.find((entry) => !entry.disabled)?.element)?.focus();

        document.addEventListener('pointerdown', onPointerDown, true);
        window.addEventListener('resize', onViewportChange);
        window.addEventListener('scroll', onViewportChange, true);
    };

    const choose = (entry: OptionEntry) => {
        if (entry.disabled) return;
        select.value = entry.value;
        sync();
        select.dispatchEvent(new Event('input', { bubbles: true }));
        select.dispatchEvent(new Event('change', { bubbles: true }));
        close(true);
    };

    trigger.addEventListener('click', () => {
        if (menu.hidden) open();
        else close(true);
    });
    trigger.addEventListener('keydown', (event) => {
        if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
            event.preventDefault();
            open(event.key === 'ArrowUp');
        }
    });

    menu.addEventListener('click', (event) => {
        const target = (event.target as Element | null)?.closest<HTMLButtonElement>('.og-select-option');
        if (!target) return;
        const entry = entries.find((item) => item.element === target);
        if (entry) choose(entry);
    });

    menu.addEventListener('keydown', (event) => {
        const options = entries.filter((entry) => !entry.disabled).map((entry) => entry.element);
        const current = options.indexOf(document.activeElement as HTMLButtonElement);
        if (event.key === 'ArrowDown') {
            event.preventDefault();
            options[(current + 1) % options.length]?.focus();
        } else if (event.key === 'ArrowUp') {
            event.preventDefault();
            options[(current - 1 + options.length) % options.length]?.focus();
        } else if (event.key === 'Home') {
            event.preventDefault();
            options[0]?.focus();
        } else if (event.key === 'End') {
            event.preventDefault();
            options[options.length - 1]?.focus();
        } else if (event.key === 'Escape') {
            event.preventDefault();
            close(true);
        } else if (event.key === 'Tab') {
            close();
        }
    });

    // Keep the visible value in step with programmatic changes and form resets.
    select.addEventListener('change', sync);
    select.form?.addEventListener('reset', () => window.setTimeout(sync, 0));
    new MutationObserver(sync).observe(select, {
        attributes: true,
        attributeFilter: ['selected', 'disabled', 'label', 'value'],
        subtree: true,
    });

    select.parentNode?.insertBefore(wrapper, select);
    wrapper.append(select, trigger, menu);
    hideNative(select);
    sync();
};

export const initOgSelects = (root: ParentNode = document) => {
    if (root instanceof Element && root.matches('select')) enhanceSelect(root as HTMLSelectElement);
    root.querySelectorAll<HTMLSelectElement>('select').forEach(enhanceSelect);
};
