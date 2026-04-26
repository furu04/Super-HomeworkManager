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

var SUBJECT_PALETTE = [
    { bg: '#4361ee', text: '#fff' },
    { bg: '#7b2d8b', text: '#fff' },
    { bg: '#2d9b4e', text: '#fff' },
    { bg: '#c87800', text: '#fff' },
    { bg: '#0077b6', text: '#fff' },
    { bg: '#c1121f', text: '#fff' },
    { bg: '#457b9d', text: '#fff' },
    { bg: '#588157', text: '#fff' },
    { bg: '#6d4c41', text: '#fff' },
    { bg: '#6a4c93', text: '#fff' },
];

function subjectColorFor(subject) {
    if (!subject) return null;
    var hash = 0;
    for (var i = 0; i < subject.length; i++) {
        hash = (hash * 31 + subject.charCodeAt(i)) | 0;
    }
    return SUBJECT_PALETTE[Math.abs(hash) % SUBJECT_PALETTE.length];
}

function applySubjectColors() {
    document.querySelectorAll('.subject-badge[data-subject]').forEach(function (badge) {
        var color = subjectColorFor(badge.dataset.subject);
        if (color) {
            badge.style.backgroundColor = color.bg;
            badge.style.color = color.text;
        }
    });
}

window.subjectColorFor = subjectColorFor;

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

    applySubjectColors();

    const dueDateInput = document.getElementById('due_date');
    if (dueDateInput && !dueDateInput.value) {
        const tomorrow = new Date();
        tomorrow.setDate(tomorrow.getDate() + 1);
        tomorrow.setHours(23, 59, 0, 0);
        dueDateInput.value = tomorrow.toISOString().slice(0, 16);
    }

    initAssignmentIndex();
});

function initAssignmentIndex() {
    if (!document.getElementById('tableView')) return;

    var _countdownInterval = null;
    var _view = localStorage.getItem('viewMode') || 'table';
    var _grouped = localStorage.getItem('grouped') === 'true';
    var _kbFocusIndex = -1;

    function getRows() {
        return Array.from(document.querySelectorAll('#tableView .assignment-row'));
    }

    function updateCountdowns() {
        var now = new Date();
        var hasUnder24h = false;

        getRows().forEach(function (row) {
            if (row.dataset.completed === 'true') return;
            var dueTs = row.dataset.dueTs;
            if (!dueTs) return;
            var due = new Date(parseInt(dueTs) * 1000);
            if (isNaN(due.getTime())) return;

            var diff = due - now;
            var countdownEl = row.querySelector('.countdown');

            row.classList.remove('anxiety-danger', 'anxiety-warning', 'bg-danger-subtle');
            if (countdownEl) countdownEl.className = 'countdown small fw-bold font-monospace';

            if (diff < 0) {
                if (countdownEl) { countdownEl.textContent = '期限切れ'; countdownEl.classList.add('text-danger'); }
                row.classList.add('bg-danger-subtle');
                return;
            }

            var days    = Math.floor(diff / 86400000);
            var hours   = Math.floor((diff % 86400000) / 3600000);
            var minutes = Math.floor((diff % 3600000) / 60000);
            var seconds = Math.floor((diff % 60000) / 1000);
            var remainingHours = days * 24 + hours;

            var text = (days > 0 ? days + '日 ' : '') +
                String(hours).padStart(2, '0') + ':' +
                String(minutes).padStart(2, '0') + ':' +
                String(seconds).padStart(2, '0');

            if (countdownEl) countdownEl.textContent = text;

            if (remainingHours < 24) {
                hasUnder24h = true;
                row.classList.add('anxiety-danger');
                if (countdownEl) {
                    countdownEl.classList.add('text-danger', 'countdown-urgent');
                    countdownEl.innerHTML = '<i class="bi bi-exclamation-triangle-fill me-1" aria-hidden="true"></i>' + text;
                }
            } else if (days < 7) {
                row.classList.add('anxiety-warning');
                if (countdownEl) {
                    countdownEl.classList.add('text-dark');
                    countdownEl.innerHTML = '<i class="bi bi-exclamation-triangle-fill text-warning me-1" aria-hidden="true"></i>' + text;
                }
            } else {
                if (countdownEl) countdownEl.classList.add('text-secondary');
            }
        });

        if (_countdownInterval !== null) return;
        var interval = hasUnder24h ? 1000 : 60000;
        _countdownInterval = setTimeout(function () { _countdownInterval = null; updateCountdowns(); }, interval);
    }

    function toggleCountdown() {
        var cols = document.querySelectorAll('.countdown-col');
        var btn = document.getElementById('toggleCountdownBtn');
        var btnText = document.getElementById('countdownBtnText');
        var isHidden = cols[0] && cols[0].style.display === 'none';
        cols.forEach(function (col) { col.style.display = isHidden ? '' : 'none'; });
        var nowHidden = !isHidden;
        btnText.textContent = nowHidden ? '残り表示' : '残り非表示';
        btn.setAttribute('aria-pressed', nowHidden ? 'true' : 'false');
        localStorage.setItem('countdownHidden', nowHidden);
    }
    window.toggleCountdown = toggleCountdown;

    function setView(mode) {
        _view = mode;
        localStorage.setItem('viewMode', mode);
        var tableView  = document.getElementById('tableView');
        var kanbanView = document.getElementById('kanbanView');
        var tableBtn   = document.getElementById('viewTableBtn');
        var kanbanBtn  = document.getElementById('viewKanbanBtn');
        var groupBtn   = document.getElementById('groupToggleBtn');

        if (mode === 'kanban') {
            tableView.classList.add('d-none');
            kanbanView.classList.remove('d-none');
            tableBtn.classList.remove('active', 'btn-secondary');
            tableBtn.classList.add('btn-outline-secondary');
            kanbanBtn.classList.remove('btn-outline-secondary');
            kanbanBtn.classList.add('active', 'btn-secondary');
            groupBtn.classList.add('d-none');
            buildKanban();
        } else {
            kanbanView.classList.add('d-none');
            tableView.classList.remove('d-none');
            kanbanBtn.classList.remove('active', 'btn-secondary');
            kanbanBtn.classList.add('btn-outline-secondary');
            tableBtn.classList.remove('btn-outline-secondary');
            tableBtn.classList.add('active', 'btn-secondary');
            groupBtn.classList.remove('d-none');
            if (_grouped) applyGrouping();
        }
    }
    window.setView = setView;

    function buildKanban() {
        var cols = { overdue: [], today: [], week: [], later: [] };
        var now = new Date();
        var startOfDay = new Date(now.getFullYear(), now.getMonth(), now.getDate());
        var endOfDay   = new Date(startOfDay.getTime() + 86400000);
        var endOfWeek  = new Date(startOfDay.getTime() + 7 * 86400000);

        getRows().forEach(function (row) {
            if (row.dataset.completed === 'true') return;
            var ts = parseInt(row.dataset.dueTs) * 1000;
            var due = new Date(ts);
            var bucket = due < now ? 'overdue' : due < endOfDay ? 'today' : due < endOfWeek ? 'week' : 'later';
            cols[bucket].push({
                id:       row.dataset.id,
                title:    row.dataset.title,
                subject:  row.dataset.subject,
                priority: row.dataset.priority,
                pinned:   row.dataset.pinned === 'true',
                dueTs:    ts
            });
        });

        var priorityColor = { high: 'danger', medium: 'warning', low: 'secondary' };
        var priorityLabel = { high: '高', medium: '中', low: '低' };
        var priorityText  = { high: 'white', medium: 'dark', low: 'white' };

        function renderCol(key, colId, countId) {
            var el = document.getElementById(colId);
            var cntEl = document.getElementById(countId);
            el.innerHTML = '';
            cntEl.textContent = cols[key].length;
            if (!cols[key].length) {
                el.innerHTML = '<p class="text-muted small text-center mt-3">なし</p>';
                return;
            }
            var csrf = (document.getElementById('_csrf_global') || {}).value || '';
            cols[key].forEach(function (a) {
                var due = new Date(a.dueTs);
                var dateStr = due.getFullYear() + '/' + String(due.getMonth() + 1).padStart(2, '0') + '/' +
                    String(due.getDate()).padStart(2, '0') + ' ' +
                    String(due.getHours()).padStart(2, '0') + ':' + String(due.getMinutes()).padStart(2, '0');
                var toggleForm = document.querySelector('form[data-row-id="' + a.id + '"]');
                var toggleAction = toggleForm ? XSS.sanitizeUrl(toggleForm.action) : '#';
                var pc = priorityColor[a.priority] || 'secondary';
                var pl = priorityLabel[a.priority] || a.priority;
                var pt = priorityText[a.priority] || 'white';
                var card = document.createElement('div');
                card.className = 'kanban-card' + (a.pinned ? ' row-pinned' : '');
                card.innerHTML =
                    '<div class="d-flex justify-content-between align-items-start mb-1">' +
                    '<div class="fw-bold" style="font-size:0.85rem;word-break:break-all;">' + XSS.escapeHtml(a.title) + '</div>' +
                    '<form action="' + toggleAction + '" method="POST" class="ms-1 flex-shrink-0">' +
                    '<input type="hidden" name="_csrf" value="' + XSS.escapeHtml(csrf) + '">' +
                    '<button type="submit" class="btn btn-sm btn-outline-success py-0 px-1" aria-label="完了にする"><i class="bi bi-check" aria-hidden="true"></i></button>' +
                    '</form></div>' +
                    '<div class="d-flex gap-1 flex-wrap">' +
                    (a.subject ? (function() {
                        var c = subjectColorFor(a.subject);
                        var bg = c ? c.bg : '#6c757d';
                        return '<span class="badge" style="font-size:0.7rem;background-color:' + bg + ';color:#fff;">' + XSS.escapeHtml(a.subject) + '</span>';
                    })() : '') +
                    '<span class="badge bg-' + pc + ' text-' + pt + '" style="font-size:0.7rem;">' + pl + '</span>' +
                    '</div>' +
                    '<div class="text-muted mt-1" style="font-size:0.75rem;">' + XSS.escapeHtml(dateStr) + '</div>' +
                    (a.pinned ? '<div class="text-warning" style="font-size:0.7rem;"><i class="bi bi-pin-fill"></i> ピン留め</div>' : '');
                el.appendChild(card);
            });
        }

        renderCol('overdue', 'kb-overdue', 'kb-count-overdue');
        renderCol('today',   'kb-today',   'kb-count-today');
        renderCol('week',    'kb-week',    'kb-count-week');
        renderCol('later',   'kb-later',   'kb-count-later');
    }

    function applyGrouping() {
        removeGrouping();
        var rows = getRows();
        if (!rows.length) return;
        var groups = {};
        var order = [];
        rows.forEach(function (row) {
            var subj = row.dataset.subject || '（科目なし）';
            if (!groups[subj]) { groups[subj] = []; order.push(subj); }
            groups[subj].push(row);
        });
        if (order.length <= 1) return;
        var tbody = rows[0].closest('tbody');
        var theadRow = rows[0].closest('table').querySelector('thead tr');
        var colCount = theadRow ? theadRow.children.length : 8;
        order.forEach(function (subj) {
            var groupRows = groups[subj];
            var headerRow = document.createElement('tr');
            headerRow.className = 'subject-group-row';
            headerRow.dataset.group = subj;
            var td = document.createElement('td');
            td.colSpan = colCount;
            td.innerHTML = '<i class="bi bi-chevron-down me-1"></i>' + XSS.escapeHtml(subj) +
                ' <span class="badge bg-secondary ms-1">' + groupRows.length + '</span>';
            headerRow.appendChild(td);
            tbody.insertBefore(headerRow, groupRows[0]);
            headerRow.addEventListener('click', function () {
                var collapsed = headerRow.classList.toggle('collapsed');
                headerRow.querySelector('i').className = collapsed ? 'bi bi-chevron-right me-1' : 'bi bi-chevron-down me-1';
                groupRows.forEach(function (r) { r.style.display = collapsed ? 'none' : ''; });
            });
        });
    }

    function removeGrouping() {
        document.querySelectorAll('.subject-group-row').forEach(function (r) { r.remove(); });
        getRows().forEach(function (r) { r.style.display = ''; });
    }

    function toggleGrouping() {
        _grouped = !_grouped;
        localStorage.setItem('grouped', _grouped);
        var btn  = document.getElementById('groupToggleBtn');
        var text = document.getElementById('groupBtnText');
        if (_grouped) {
            applyGrouping();
            btn.classList.add('btn-secondary', 'text-white');
            btn.classList.remove('btn-outline-secondary');
            text.textContent = 'グループ解除';
        } else {
            removeGrouping();
            btn.classList.remove('btn-secondary', 'text-white');
            btn.classList.add('btn-outline-secondary');
            text.textContent = 'グループ化';
        }
    }
    window.toggleGrouping = toggleGrouping;

    function updateBulkBar() {
        var checked = document.querySelectorAll('.row-check:checked');
        var bar     = document.getElementById('bulkBar');
        var countEl = document.getElementById('bulkCount');
        if (checked.length > 0) {
            bar.classList.remove('d-none');
            countEl.textContent = checked.length + '件選択中';
        } else {
            bar.classList.add('d-none');
        }
    }

    function getCheckedIDs() {
        return Array.from(document.querySelectorAll('.row-check:checked')).map(function (c) { return c.value; });
    }

    function submitBulkComplete() {
        var ids = getCheckedIDs();
        if (!ids.length) return;
        var form = document.getElementById('bulkCompleteForm');
        ids.forEach(function (id) {
            var inp = document.createElement('input');
            inp.type = 'hidden'; inp.name = 'ids'; inp.value = id;
            form.appendChild(inp);
        });
        form.submit();
    }
    window.submitBulkComplete = submitBulkComplete;

    function confirmBulkDelete() {
        var ids = getCheckedIDs();
        if (!ids.length) return;

        var recurringMap = {};
        ids.forEach(function (id) {
            var row = document.querySelector('.assignment-row[data-id="' + id + '"]');
            if (row && row.dataset.recurringId) {
                var rid = row.dataset.recurringId;
                if (!recurringMap[rid]) {
                    recurringMap[rid] = { title: row.dataset.title, count: 0 };
                }
                recurringMap[rid].count++;
            }
        });

        var recurringKeys = Object.keys(recurringMap);

        if (recurringKeys.length > 0) {
            var list = document.getElementById('bulkDeleteRecurringList');
            list.innerHTML = '';
            recurringKeys.forEach(function (rid) {
                var item = recurringMap[rid];
                var li = document.createElement('li');
                li.className = 'list-group-item py-2 px-2 small';
                li.innerHTML = '<i class="bi bi-repeat text-info me-2" aria-hidden="true"></i>' +
                    XSS.escapeHtml(item.title) +
                    (item.count > 1 ? ' <span class="badge bg-secondary ms-1">' + item.count + '件</span>' : '');
                list.appendChild(li);
            });

            var modalEl = document.getElementById('bulkDeleteRecurringModal');
            var modal = new bootstrap.Modal(modalEl);

            document.getElementById('bulkDeleteOnlyBtn').onclick = function () {
                modal.hide();
                submitBulkDeleteForm(ids, false);
            };
            document.getElementById('bulkDeleteWithRecurringBtn').onclick = function () {
                modal.hide();
                submitBulkDeleteForm(ids, true);
            };

            modal.show();
        } else {
            showConfirmModal(ids.length + '件の課題を削除しますか？', function () {
                submitBulkDeleteForm(ids, false);
            });
        }
    }
    window.confirmBulkDelete = confirmBulkDelete;

    function submitBulkDeleteForm(ids, deleteRecurring) {
        var form = document.getElementById('bulkDeleteForm');
        form.querySelectorAll('input[name="ids"], input[name="delete_recurring"]').forEach(function (inp) { inp.remove(); });
        ids.forEach(function (id) {
            var inp = document.createElement('input');
            inp.type = 'hidden'; inp.name = 'ids'; inp.value = id;
            form.appendChild(inp);
        });
        if (deleteRecurring) {
            var inp = document.createElement('input');
            inp.type = 'hidden'; inp.name = 'delete_recurring'; inp.value = 'true';
            form.appendChild(inp);
        }
        form.submit();
    }

    function clearSelection() {
        document.querySelectorAll('.row-check, #selectAll').forEach(function (c) { c.checked = false; });
        updateBulkBar();
    }
    window.clearSelection = clearSelection;

    function moveFocus(delta) {
        var rows = getRows().filter(function (r) { return r.style.display !== 'none'; });
        if (!rows.length) return;
        rows.forEach(function (r) { r.classList.remove('kb-focus'); });
        _kbFocusIndex = Math.max(0, Math.min(rows.length - 1, _kbFocusIndex + delta));
        rows[_kbFocusIndex].classList.add('kb-focus');
        rows[_kbFocusIndex].scrollIntoView({ block: 'nearest' });
    }

    function toggleFocused() {
        var rows = getRows().filter(function (r) { return r.style.display !== 'none'; });
        if (_kbFocusIndex < 0 || _kbFocusIndex >= rows.length) return;
        var form = rows[_kbFocusIndex].querySelector('form[data-row-id]');
        if (form) form.submit();
    }

    document.addEventListener('keydown', function (e) {
        if (!document.getElementById('tableView')) return;
        if (['INPUT', 'TEXTAREA', 'SELECT'].includes(e.target.tagName)) return;
        if (e.ctrlKey || e.metaKey || e.altKey) return;
        switch (e.key) {
            case '/':
                e.preventDefault();
                var s = document.getElementById('searchInput');
                if (s) s.focus();
                break;
            case 'j': moveFocus(1); break;
            case 'k': moveFocus(-1); break;
            case 'x': toggleFocused(); break;
            case 'n':
                if (!document.activeElement || document.activeElement === document.body) {
                    window.location.href = '/assignments/new';
                }
                break;
            case 'Escape': clearSelection(); break;
        }
    });

    var selectAll = document.getElementById('selectAll');
    if (selectAll) {
        selectAll.addEventListener('change', function () {
            document.querySelectorAll('.row-check').forEach(function (c) { c.checked = selectAll.checked; });
            updateBulkBar();
        });
    }
    document.querySelectorAll('.row-check').forEach(function (c) {
        c.addEventListener('change', function () {
            var all     = document.querySelectorAll('.row-check');
            var checked = document.querySelectorAll('.row-check:checked');
            if (selectAll) selectAll.checked = all.length === checked.length;
            updateBulkBar();
        });
    });

    var recurringModal = document.getElementById('recurringModal');
    if (recurringModal) {
        recurringModal.addEventListener('show.bs.modal', function (event) {
            var button   = event.relatedTarget;
            var id       = button.getAttribute('data-recurring-id');
            var title    = button.getAttribute('data-recurring-title');
            var type     = button.getAttribute('data-recurring-type');
            var isActive = button.getAttribute('data-recurring-active') === 'true';
            document.getElementById('recurringModalTitle').textContent = title;
            document.getElementById('recurringStopForm').action = '/recurring/' + id + '/stop';
            document.getElementById('recurringEditBtn').href = '/recurring/' + id + '/edit';
            var typeLabels = { daily: '毎日', weekly: '毎週', monthly: '毎月', unknown: '(不明)' };
            document.getElementById('recurringTypeLabel').textContent = typeLabels[type] || type || '不明';
            var statusEl = document.getElementById('recurringStatus');
            if (isActive) {
                statusEl.innerHTML = '<span class="badge bg-success">有効</span>';
                document.getElementById('recurringStopBtn').style.display = 'inline-block';
            } else {
                statusEl.innerHTML = '<span class="badge bg-secondary">停止中</span>';
                document.getElementById('recurringStopBtn').style.display = 'none';
            }
        });
    }

    if (localStorage.getItem('countdownHidden') === 'true') {
        document.querySelectorAll('.countdown-col').forEach(function (col) { col.style.display = 'none'; });
        var btn     = document.getElementById('toggleCountdownBtn');
        var btnText = document.getElementById('countdownBtnText');
        if (btnText) btnText.textContent = '残り表示';
        if (btn) btn.setAttribute('aria-pressed', 'true');
    }

    var gBtn  = document.getElementById('groupToggleBtn');
    var gText = document.getElementById('groupBtnText');
    if (_grouped && gBtn) {
        gBtn.classList.add('btn-secondary', 'text-white');
        gBtn.classList.remove('btn-outline-secondary');
        if (gText) gText.textContent = 'グループ解除';
    }

    window.showDeleteRecurringModal = function (assignmentId, recurringId) {
        var modal = new bootstrap.Modal(document.getElementById('deleteRecurringModal'));
        document.getElementById('deleteOnlyForm').action = '/assignments/' + assignmentId + '/delete';
        document.getElementById('deleteAndStopForm').action = '/assignments/' + assignmentId + '/delete?stop_recurring=' + recurringId;
        modal.show();
    };

    setView(_view);
    updateCountdowns();
}
