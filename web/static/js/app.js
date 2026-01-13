const XSS = {
    escapeHtml: function (str) {
        if (str === null || str === undefined) return '';
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    },

    setTextSafe: function (element, text) {
        if (element) {
            element.textContent = text;
        }
    },

    sanitizeUrl: function (url) {
        if (!url) return '';
        const cleaned = String(url).replace(/[\x00-\x1F\x7F]/g, '').trim();
        try {
            const parsed = new URL(cleaned, window.location.origin);
            if (parsed.protocol === 'http:' || parsed.protocol === 'https:') {
                return parsed.href;
            }
        } catch (e) {
            if (cleaned.startsWith('/') && !cleaned.startsWith('//')) {
                return cleaned;
            }
        }
        return '';
    }
};

window.XSS = XSS;

document.addEventListener('DOMContentLoaded', function () {
    const alerts = document.querySelectorAll('.alert:not(.alert-danger):not(.modal .alert)');
    alerts.forEach(function (alert) {
        setTimeout(function () {
            alert.classList.add('fade');
            setTimeout(function () {
                alert.remove();
            }, 150);
        }, 5000);
    });

    const confirmForms = document.querySelectorAll('form[data-confirm]');
    confirmForms.forEach(function (form) {
        form.addEventListener('submit', function (e) {
            if (!confirm(form.dataset.confirm)) {
                e.preventDefault();
            }
        });
    });

    const dueDateInput = document.getElementById('due_date');
    if (dueDateInput && !dueDateInput.value) {
        const tomorrow = new Date();
        tomorrow.setDate(tomorrow.getDate() + 1);
        tomorrow.setHours(23, 59, 0, 0);
        dueDateInput.value = tomorrow.toISOString().slice(0, 16);
    }
});
