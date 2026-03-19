(function () {
    function q(id) {
        return document.getElementById(id);
    }

    function getByPath(obj, path) {
        if (!obj || !path) return undefined;
        const parts = String(path).split('.');
        let current = obj;
        for (const part of parts) {
            if (current == null) return undefined;
            current = current[part];
        }
        return current;
    }

    function t(key, fallback) {
        const messages = window.__CLAWGET_I18N_MESSAGES__;
        const value = getByPath(messages, key);
        if (typeof value === 'string' && value) return value;
        return fallback || '';
    }

    function currentLang() {
        return document.documentElement.lang || 'zh-CN';
    }

    function appendParams(baseUrl, params) {
        try {
            const url = new URL(baseUrl, window.location.href);
            Object.entries(params).forEach(([k, v]) => {
                if (v == null || v === '') return;
                url.searchParams.set(k, String(v));
            });
            return url.toString();
        } catch {
            return baseUrl;
        }
    }

    function isValidEmail(email) {
        if (!email) return false;
        const value = String(email).trim();
        if (!value) return false;
        if (value.length > 254) return false;
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
    }

    const backdrop = q('payModalBackdrop');
    const closeBtn = q('payModalClose');
    const cancelBtn = q('payCancel');
    const form = q('payEmailForm');
    const emailInput = q('payEmail');
    const statusEl = q('payModalStatus');

    if (!backdrop || !form || !emailInput || !statusEl) return;

    let plan = 'basic';

    function isOpen() {
        return backdrop.classList.contains('open');
    }

    function setStatus(key, fallback) {
        statusEl.textContent = t(key, fallback);
    }

    function openModal(nextPlan) {
        plan = nextPlan || 'basic';
        backdrop.classList.add('open');
        backdrop.setAttribute('aria-hidden', 'false');
        setStatus('payment.modal.hint', currentLang() === 'en' ? 'Enter your email to continue' : '请输入邮箱后继续');
        window.setTimeout(() => emailInput.focus(), 0);
    }

    function closeModal() {
        backdrop.classList.remove('open');
        backdrop.setAttribute('aria-hidden', 'true');
    }

    function resolvePayPageUrl(planKey) {
        const key = planKey === 'pro' ? 'payment.pages.pro' : 'payment.pages.basic';
        return t(key, './pay.html');
    }

    function goToPayPage(emailValue) {
        const base = resolvePayPageUrl(plan);
        const url = appendParams(base, { plan, email: emailValue, lang: currentLang() });
        window.location.href = url;
    }

    document.addEventListener('click', (e) => {
        const target = e.target;
        const btn = target?.closest?.('.pay-open');
        if (btn) {
            e.preventDefault();
            const nextPlan = btn.getAttribute('data-plan') || 'basic';
            openModal(nextPlan);
            return;
        }

        if (target === backdrop) closeModal();
    });

    closeBtn?.addEventListener('click', closeModal);
    cancelBtn?.addEventListener('click', closeModal);

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && isOpen()) closeModal();
    });

    const htmlLangObserver = new MutationObserver(() => {
        if (!isOpen()) return;
        setStatus('payment.modal.hint');
    });
    htmlLangObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });

    form.addEventListener('submit', (e) => {
        e.preventDefault();
        const emailValue = String(emailInput.value || '').trim();
        if (!emailValue) {
            setStatus('payment.status.emailRequired', currentLang() === 'en' ? 'Please enter your email.' : '请输入邮箱。');
            emailInput.focus();
            return;
        }
        if (!isValidEmail(emailValue)) {
            setStatus('payment.status.emailInvalid', currentLang() === 'en' ? 'Invalid email format.' : '邮箱格式不正确。');
            emailInput.focus();
            return;
        }
        goToPayPage(emailValue);
    });
})();
