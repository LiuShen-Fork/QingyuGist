// Top-right flash notifications. Each toast owns its timer and progress bar, so
// hovering it pauses the countdown instead of dismissing something mid-read.
const initToast = (toast: HTMLElement) => {
    if (toast.dataset.toastReady === '1') return;
    toast.dataset.toastReady = '1';

    const duration = Number.parseFloat(toast.style.getPropertyValue('--toast-duration')) || 4500;
    const progress = toast.querySelector<HTMLElement>('.og-toast-progress');
    let remaining = duration;
    let startedAt = 0;
    let timer: number | undefined;

    const dismiss = () => {
        window.clearTimeout(timer);
        toast.classList.add('og-toast-out');
        window.setTimeout(() => toast.remove(), 200);
    };

    toast.querySelector<HTMLElement>('[data-toast-close]')?.addEventListener('click', dismiss);

    const start = () => {
        startedAt = Date.now();
        window.clearTimeout(timer);
        timer = window.setTimeout(dismiss, remaining);
        if (progress) progress.style.animationPlayState = 'running';
    };

    toast.addEventListener('mouseenter', () => {
        remaining = Math.max(0, remaining - (Date.now() - startedAt));
        window.clearTimeout(timer);
        if (progress) progress.style.animationPlayState = 'paused';
    });
    toast.addEventListener('mouseleave', start);

    start();
};

export const initToasts = (root: ParentNode = document) => {
    root.querySelectorAll<HTMLElement>('[data-toast]').forEach(initToast);
};
