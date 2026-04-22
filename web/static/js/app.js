const XSS = {
    escapeHtml: function (str) {
        if (str === null || str === undefined) return '';
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    },

    setTextSafe: function (element, text) {
        if (element) element.textContent = text;
    },

    sanitizeUrl: function (url) {
        if (!url) return '';
        const cleaned = String(url).replace(/[\x00-\x1F\x7F]/g, '').trim();
        try {
            const parsed = new URL(cleaned, window.location.origin);
            if (parsed.protocol === 'http:' || parsed.protocol === 'https:') return parsed.href;
        } catch (e) {
            if (cleaned.startsWith('/') && !cleaned.startsWith('//')) return cleaned;
        }
        return '';
    }
};

window.XSS = XSS;

let _pendingConfirmForm = null;

function showConfirmModal(message, onOk) {
    const bodyEl = document.getElementById('confirmModalBody');
    const okBtn  = document.getElementById('confirmModalOk');
    if (!bodyEl || !okBtn) { if (onOk) onOk(); return; }
    bodyEl.textContent = message;
    const handler = function () {
        okBtn.removeEventListener('click', handler);
        bootstrap.Modal.getInstance(document.getElementById('confirmModal')).hide();
        if (onOk) onOk();
    };
    okBtn.addEventListener('click', handler);
    new bootstrap.Modal(document.getElementById('confirmModal')).show();
}

window.showConfirmModal = showConfirmModal;

function setupFormSubmitOnce(form) {
    form.addEventListener('submit', function () {
        const btn = form.querySelector('[type=submit]');
        if (!btn || btn.disabled) return;
        btn.disabled = true;
        const orig = btn.innerHTML;
        btn.innerHTML = '<span class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>処理中...';
        window.addEventListener('pageshow', function () {
            btn.disabled = false;
            btn.innerHTML = orig;
        }, { once: true });
    });
}

function showCopyFeedback(message) {
    let el = document.getElementById('globalCopyFeedback');
    if (!el) {
        el = document.createElement('div');
        el.id = 'globalCopyFeedback';
        el.className = 'copy-feedback alert alert-success shadow-sm py-2 px-3';
        document.body.appendChild(el);
    }
    el.textContent = message;
    el.classList.add('show');
    clearTimeout(el._timeout);
    el._timeout = setTimeout(function () { el.classList.remove('show'); }, 2000);
}

window.showCopyFeedback = showCopyFeedback;

document.addEventListener('DOMContentLoaded', function () {
    const alerts = document.querySelectorAll('.alert:not(.alert-danger):not(.modal .alert)');
    alerts.forEach(function (alert) {
        setTimeout(function () {
            alert.classList.add('fade');
            setTimeout(function () { alert.remove(); }, 150);
        }, 5000);
    });

    document.querySelectorAll('form[data-confirm]').forEach(function (form) {
        form.addEventListener('submit', function (e) {
            e.preventDefault();
            const msg = form.dataset.confirm;
            showConfirmModal(msg, function () { form.submit(); });
        });
    });

    document.querySelectorAll('form:not([data-confirm])').forEach(setupFormSubmitOnce);

    const dueDateInput = document.getElementById('due_date');
    if (dueDateInput && !dueDateInput.value) {
        const tomorrow = new Date();
        tomorrow.setDate(tomorrow.getDate() + 1);
        tomorrow.setHours(23, 59, 0, 0);
        dueDateInput.value = tomorrow.toISOString().slice(0, 16);
    }
});
