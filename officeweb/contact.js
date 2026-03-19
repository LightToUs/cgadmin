(function () {
    function q(id) {
        return document.getElementById(id);
    }

    const backdrop = q('contactModalBackdrop');
    const closeBtn = q('contactModalClose');
    const cancelBtn = q('contactCancel');
    const form = q('contactForm');
    const statusEl = q('contactStatus');
    const sendBtn = q('contactSend');
    const configEl = q('contactConfig');

    if (!backdrop || !form || !configEl) return;

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

    function isOpen() {
        return backdrop.classList.contains('open');
    }

    function openModal() {
        backdrop.classList.add('open');
        backdrop.setAttribute('aria-hidden', 'false');
        statusEl.textContent = t('contact.modal.hint', statusEl.textContent);
        window.setTimeout(() => {
            q('contactName')?.focus();
        }, 0);
    }

    function closeModal() {
        backdrop.classList.remove('open');
        backdrop.setAttribute('aria-hidden', 'true');
    }

    function setBusy(busy) {
        sendBtn.disabled = busy;
        sendBtn.style.opacity = busy ? '0.7' : '1';
    }

    function setStatus(textKey, fallback) {
        statusEl.textContent = t(textKey, fallback);
    }

    function getLang() {
        return document.documentElement.lang || 'zh-CN';
    }

    function getEmailJsConfig() {
        const publicKey = (configEl.getAttribute('data-public') || '').trim();
        const serviceId = (configEl.getAttribute('data-service') || '').trim();
        const templateId = (configEl.getAttribute('data-template') || '').trim();
        const to = (configEl.getAttribute('data-to') || '').trim();
        return { publicKey, serviceId, templateId, to };
    }

    async function sendEmail(params) {
        const { publicKey, serviceId, templateId } = getEmailJsConfig();
        if (!publicKey || !serviceId || !templateId) {
            throw new Error('missing_emailjs_config');
        }
        if (!window.emailjs || typeof window.emailjs.send !== 'function') {
            throw new Error('missing_emailjs_lib');
        }
        return window.emailjs.send(serviceId, templateId, params, { publicKey });
    }

    document.addEventListener('click', (e) => {
        const target = e.target;
        const openBtn = target?.closest?.('.contact-open');
        if (openBtn) {
            e.preventDefault();
            openModal();
            return;
        }

        if (target === backdrop) {
            closeModal();
        }
    });

    closeBtn?.addEventListener('click', closeModal);
    cancelBtn?.addEventListener('click', closeModal);

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && isOpen()) closeModal();
    });

    const htmlLangObserver = new MutationObserver(() => {
        if (!isOpen()) return;
        setStatus('contact.modal.hint');
    });
    htmlLangObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const name = (q('contactName')?.value || '').trim();
        const email = (q('contactEmail')?.value || '').trim();
        const message = (q('contactMessage')?.value || '').trim();

        if (!message) {
            setStatus('contact.status.messageRequired', getLang() === 'en' ? 'Please enter a message.' : '请填写内容。');
            q('contactMessage')?.focus();
            return;
        }

        setBusy(true);
        setStatus('contact.status.sending', getLang() === 'en' ? 'Sending…' : '发送中…');

        const { to } = getEmailJsConfig();
        const subject = getLang() === 'en' ? 'Contact from website' : '官网联系';

        try {
            await sendEmail({
                from_name: name || (getLang() === 'en' ? 'Anonymous' : '匿名'),
                reply_to: email,
                message,
                to_email: to,
                subject
            });

            setStatus('contact.status.sent', getLang() === 'en' ? 'Sent successfully.' : '发送成功。');
            form.reset();
            window.setTimeout(() => closeModal(), 800);
        } catch (err) {
            const msg = String(err?.message || '');
            if (msg === 'missing_emailjs_config') {
                setStatus('contact.status.notConfigured', getLang() === 'en' ? 'Email service not configured.' : '邮件服务未配置。');
            } else {
                setStatus('contact.status.failed', getLang() === 'en' ? 'Failed to send. Please try again.' : '发送失败，请稍后重试。');
            }
        } finally {
            setBusy(false);
        }
    });
})();
