(function () {
    const supportedLangs = ['zh-CN', 'zh-HK', 'en'];
    const loadedLangs = new Set();

    function normalizeLang(rawLang) {
        if (!rawLang) return null;
        const lang = String(rawLang).trim();
        if (!lang) return null;
        if (supportedLangs.includes(lang)) return lang;
        const lower = lang.toLowerCase();
        if (lower.startsWith('zh-hk') || lower.startsWith('zh-tw')) return 'zh-HK';
        if (lower.startsWith('zh')) return 'zh-CN';
        if (lower.startsWith('en')) return 'en';
        return null;
    }

    function getQueryLang() {
        try {
            const url = new URL(window.location.href);
            return url.searchParams.get('lang');
        } catch {
            return null;
        }
    }

    function setQueryLang(lang) {
        try {
            const url = new URL(window.location.href);
            if (lang === 'en') {
                url.searchParams.set('lang', 'en');
            } else {
                url.searchParams.delete('lang');
            }
            window.history.replaceState({}, '', url.toString());
        } catch {
        }
    }

    function detectInitialLang() {
        const fromQuery = normalizeLang(getQueryLang());
        if (fromQuery) return fromQuery;

        const fromStorage = normalizeLang(window.localStorage.getItem('lang'));
        if (fromStorage) return fromStorage;

        return 'zh-CN';
    }

    function getByPath(obj, path) {
        if (!obj) return undefined;
        if (!path) return undefined;
        const parts = String(path).split('.');
        let current = obj;
        for (const part of parts) {
            if (current == null) return undefined;
            current = current[part];
        }
        return current;
    }

    function applyTextTranslations(messages) {
        const titleKey = document.querySelector('meta[name="i18n-title"]')?.getAttribute('content');
        const title = typeof titleKey === 'string' && titleKey ? getByPath(messages, titleKey) : getByPath(messages, 'meta.title');
        if (typeof title === 'string') document.title = title;

        const lang = getByPath(messages, 'lang');
        if (typeof lang === 'string') document.documentElement.lang = lang;

        document.querySelectorAll('[data-i18n]').forEach((el) => {
            const key = el.getAttribute('data-i18n');
            const value = getByPath(messages, key);
            if (typeof value === 'string') el.textContent = value;
        });

        document.querySelectorAll('[data-i18n-html]').forEach((el) => {
            const key = el.getAttribute('data-i18n-html');
            const value = getByPath(messages, key);
            if (typeof value === 'string') el.innerHTML = value;
        });

        document.querySelectorAll('[data-i18n-attr]').forEach((el) => {
            const raw = el.getAttribute('data-i18n-attr') || '';
            raw.split(',').map(s => s.trim()).filter(Boolean).forEach((pair) => {
                const idx = pair.indexOf(':');
                if (idx <= 0) return;
                const attr = pair.slice(0, idx).trim();
                const key = pair.slice(idx + 1).trim();
                const value = getByPath(messages, key);
                if (typeof value === 'string') el.setAttribute(attr, value);
            });
        });
    }

    function wireCopyButtons(t) {
        document.querySelectorAll('.copy-btn').forEach((btn) => {
            btn.addEventListener('click', async () => {
                const copyText = btn.getAttribute('data-copy-text') || '';
                try {
                    if (navigator.clipboard && navigator.clipboard.writeText) {
                        await navigator.clipboard.writeText(copyText);
                    }
                } catch {
                }

                const originalText = btn.innerText;
                btn.innerText = t('common.copied') || originalText;
                window.setTimeout(() => {
                    btn.innerText = originalText;
                }, 2000);
            });
        });
    }

    async function loadMessages(lang) {
        const store = window.__CLAWGET_I18N__;
        if (store && store[lang]) return store[lang];
        if (loadedLangs.has(lang)) {
            await new Promise((resolve) => window.setTimeout(resolve, 0));
            if (window.__CLAWGET_I18N__ && window.__CLAWGET_I18N__[lang]) return window.__CLAWGET_I18N__[lang];
        }

        loadedLangs.add(lang);

        await new Promise((resolve, reject) => {
            const script = document.createElement('script');
            script.async = true;
            script.src = `./i18n/index.${lang}.js`;
            script.onload = () => resolve();
            script.onerror = () => reject(new Error(`Failed to load i18n script: ${lang}`));
            document.head.appendChild(script);
        });

        if (window.__CLAWGET_I18N__ && window.__CLAWGET_I18N__[lang]) return window.__CLAWGET_I18N__[lang];
        throw new Error(`Missing i18n messages after loading: ${lang}`);
    }

    function renderLangSelect(messages, currentLang) {
        const select = document.getElementById('langSelect');
        if (!select) return;

        const options = supportedLangs.map((lang) => {
            const label = getByPath(messages, `common.langNames.${lang}`) || lang;
            const selected = lang === currentLang ? ' selected' : '';
            return `<option value="${lang}"${selected}>${label}</option>`;
        });

        select.innerHTML = options.join('');
    }

    function setCurrentI18n(messages, lang) {
        window.__CLAWGET_I18N_MESSAGES__ = messages;
        window.__CLAWGET_I18N_LANG__ = lang;
    }

    async function init() {
        const fromQuery = normalizeLang(getQueryLang());
        let currentLang = detectInitialLang();
        let messages;

        try {
            messages = await loadMessages(currentLang);
        } catch {
            currentLang = 'zh-CN';
            messages = await loadMessages(currentLang);
        }

        if (fromQuery) {
            window.localStorage.setItem('lang', currentLang);
        }

        const t = (key) => {
            const value = getByPath(messages, key);
            return typeof value === 'string' ? value : '';
        };

        renderLangSelect(messages, currentLang);
        applyTextTranslations(messages);
        setCurrentI18n(messages, currentLang);

        const select = document.getElementById('langSelect');
        if (select) {
            select.addEventListener('change', async () => {
                const next = normalizeLang(select.value) || 'zh-CN';
                window.localStorage.setItem('lang', next);
                setQueryLang(next);
                messages = await loadMessages(next);
                renderLangSelect(messages, next);
                applyTextTranslations(messages);
                setCurrentI18n(messages, next);
            });
        }

        wireCopyButtons(t);
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
