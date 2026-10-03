import React, { useState, useEffect } from 'react';
import { useI18n } from '../i18n';
import { GxSaveConfig, GxTestConnection, GxSelectFolder, GxGetDefaultOutputPath, GxListModels, GxGetLogPath, GxReadLog, GxClearLog, GxOpenFolderOfFile } from '../../wailsjs/go/main/App';
import { HelixIcons } from './IconLibrary';

const Settings = ({ config, onSave }) => {
    const { t, lang, changeLang } = useI18n();
    const [localConfig, setLocalConfig] = useState({ ...config });
    const [saving, setSaving] = useState(false);
    const [showKey, setShowKey] = useState(false);
    const [testStatus, setTestStatus] = useState('');
    const [defaultPath, setDefaultPath] = useState('');
    const [availableModels, setAvailableModels] = useState([]);
    const [loadingModels, setLoadingModels] = useState(false);
    const [showLogs, setShowLogs] = useState(false);
    const [logContent, setLogContent] = useState('');
    const [logPath, setLogPath] = useState('');
    const [loadingLogs, setLoadingLogs] = useState(false);
    const [logsCopied, setLogsCopied] = useState(false);

    useEffect(() => {
        const fetchDefault = async () => {
            const def = await GxGetDefaultOutputPath();
            setDefaultPath(def);
        };
        fetchDefault();
    }, []);

    useEffect(() => {
        const fetchModels = async () => {
            if (!localConfig.api_key) return;

            setLoadingModels(true);
            try {
                // Pass current local API key and model to list models
                const models = await GxListModels(localConfig.api_key, localConfig.model);
                if (models && models.length > 0) {
                    setAvailableModels(models);
                }
            } catch (err) {
                console.error('Failed to fetch models:', err);
            } finally {
                setLoadingModels(false);
            }
        };

        // Fetch models when API key changes (with a slight delay to avoid excessive calls)
        const timer = setTimeout(() => {
            fetchModels();
        }, 500);

        return () => clearTimeout(timer);
    }, [localConfig.api_key]);

    const handleSave = async () => {
        setSaving(true);
        const err = await GxSaveConfig(localConfig);
        if (err) {
            alert(err);
        } else {
            onSave(localConfig);
        }
        setSaving(false);
    };

    const handleTestConnection = async () => {
        setTestStatus('testing');
        try {
            // Pass current local API key and model to test connection
            const result = await GxTestConnection(localConfig.api_key, localConfig.model);
            if (result) {
                // Success case
                console.log('Test connection success:', result);
                setTestStatus('success');
                // Keep success status visible for 5 seconds
                setTimeout(() => setTestStatus(''), 5000);

                // If successful, also refresh models manually to be sure
                const models = await GxListModels(localConfig.api_key, localConfig.model);
                if (models && models.length > 0) {
                    setAvailableModels(models);
                }
            } else {
                throw new Error('No response from server');
            }
        } catch (err) {
            console.error('Test connection error:', err);
            setTestStatus('error');
            // Keep error status visible for 5 seconds
            setTimeout(() => setTestStatus(''), 5000);
        }
    };

    const refreshLogs = async () => {
        setLoadingLogs(true);
        try {
            const [path, content] = await Promise.all([GxGetLogPath(), GxReadLog(200)]);
            setLogPath(path || '');
            setLogContent(content || '');
        } catch (err) {
            setLogContent(`Unable to read log: ${err}`);
        } finally {
            setLoadingLogs(false);
        }
    };

    const handleToggleLogs = async () => {
        const nextShowLogs = !showLogs;
        setShowLogs(nextShowLogs);
        if (nextShowLogs) {
            await refreshLogs();
        }
    };

    const handleCopyLogs = async () => {
        await navigator.clipboard.writeText(logContent);
        setLogsCopied(true);
        setTimeout(() => setLogsCopied(false), 2000);
    };

    const handleClearLogs = async () => {
        if (!window.confirm(t('settings.confirmClearLogs'))) return;
        await GxClearLog();
        setLogContent('');
    };

    const handleBrowse = async () => {
        try {
            // Use current path or default if empty to start the picker
            const currentPath = localConfig.output_path || defaultPath;
            const folder = await GxSelectFolder(currentPath);
            if (folder) {
                setLocalConfig({ ...localConfig, output_path: folder });
            }
        } catch (err) {
            alert(err);
        }
    };

    return (
        <main className="flex-1 flex flex-col items-center py-10 px-4 md:px-10 lg:px-40 overflow-y-auto bg-background-light dark:bg-background-dark">
            <div className="max-w-[960px] w-full flex flex-col gap-8">
                <div className="flex flex-wrap justify-between gap-3 px-4">
                    <div className="flex min-w-72 flex-col gap-3">
                        <h1 className="text-4xl font-black leading-tight tracking-tight text-slate-900 dark:text-white">{t('settings.title')}</h1>
                        <p className="text-text-muted dark:text-text-secondary text-base font-normal leading-normal max-w-2xl">
                            {t('settings.subtitle')}
                        </p>
                    </div>
                </div>

                {/* AI Section */}
                <section className="flex flex-col gap-4">
                    <div className="px-4 pb-2 pt-4 border-b border-border-light dark:border-border-dark flex items-center gap-3">
                        <span className="material-symbols-outlined text-primary">psychology</span>
                        <h2 className="text-xl font-bold leading-tight tracking-tight">{t('settings.aiSection')}</h2>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6 px-4 py-2">
                        <label className="flex flex-col flex-1 gap-2">
                            <p className="text-base font-medium leading-normal">{t('settings.provider')}</p>
                            <div className="relative">
                                <select
                                    disabled
                                    className="w-full appearance-none rounded-lg border border-border-light dark:border-border-dark bg-surface-light dark:bg-surface-dark px-4 h-14 text-base focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all cursor-not-allowed"
                                >
                                    <option value="google">Google Gemini API</option>
                                </select>
                                <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-text-muted">
                                    <span className="material-symbols-outlined">expand_more</span>
                                </div>
                            </div>
                            <p className="text-xs text-text-muted mt-1 flex items-center gap-1">
                                <span className="material-symbols-outlined text-[14px]">info</span>
                                Using ai.google.dev API with API key authentication
                            </p>
                        </label>

                        <label className="flex flex-col flex-1 gap-2">
                            <p className="text-base font-medium leading-normal">{t('settings.model')}</p>
                            <div className="relative">
                                <select
                                    value={localConfig.model}
                                    onChange={(e) => setLocalConfig({ ...localConfig, model: e.target.value })}
                                    disabled={loadingModels}
                                    className="w-full appearance-none rounded-lg border border-border-light dark:border-border-dark bg-surface-light dark:bg-surface-dark px-4 h-14 text-base focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {loadingModels ? (
                                        <option>{t('settings.loadingModels') || 'Loading models...'}</option>
                                    ) : availableModels.length > 0 ? (
                                        availableModels.map(model => {
                                            // Clean up model name for display (remove "models/" prefix)
                                            const displayName = model.replace('models/', '');
                                            return <option key={model} value={displayName}>{displayName}</option>
                                        })
                                    ) : (
                                        <>
                                            <option value="gemini-2.5-flash">Gemini 2.5 Flash</option>
                                            <option value="gemini-2.5-pro">Gemini 2.5 Pro</option>
                                            <option value="gemini-3-flash-preview">Gemini 3 Flash (Preview)</option>
                                        </>
                                    )}
                                </select>
                                <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-text-muted">
                                    <span className={`material-symbols-outlined ${loadingModels ? 'animate-spin' : ''}`}>
                                        {loadingModels ? 'sync' : 'expand_more'}
                                    </span>
                                </div>
                            </div>
                        </label>
                    </div>

                    <div className="px-4 py-2">
                        <label className="flex flex-col flex-1 gap-2">
                            <p className="text-base font-medium leading-normal">{t('settings.apiKey')}</p>
                            <div className="flex flex-col md:flex-row gap-3">
                                <div className="relative flex-1">
                                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                        <span className="material-symbols-outlined text-text-muted">key</span>
                                    </div>
                                    <input
                                        type={showKey ? "text" : "password"}
                                        value={localConfig.api_key}
                                        onChange={(e) => setLocalConfig({ ...localConfig, api_key: e.target.value })}
                                        className="w-full rounded-lg border border-border-light dark:border-border-dark bg-surface-light dark:bg-surface-dark pl-10 pr-12 h-14 text-base focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all"
                                        placeholder="Entrez votre clé API..."
                                    />
                                    <button
                                        onClick={() => setShowKey(!showKey)}
                                        className="absolute inset-y-0 right-0 pr-3 flex items-center text-text-muted hover:text-primary transition-colors"
                                    >
                                        <span className="material-symbols-outlined">
                                            {showKey ? 'visibility' : 'visibility_off'}
                                        </span>
                                    </button>
                                </div>
                                <button
                                    onClick={handleTestConnection}
                                    disabled={testStatus === 'testing'}
                                    className={`flex items-center justify-center gap-2 h-14 px-6 rounded-lg border font-medium whitespace-nowrap transition-all disabled:opacity-50 ${testStatus === 'success'
                                        ? 'bg-green-500/10 border-green-500 text-green-500'
                                        : testStatus === 'error'
                                            ? 'bg-red-500/10 border-red-500 text-red-500'
                                            : 'bg-surface-light dark:bg-surface-dark border-border-light dark:border-border-dark hover:border-primary hover:text-primary'
                                        }`}
                                >
                                    <span className={`material-symbols-outlined ${testStatus === 'testing' ? 'animate-spin' : ''}`}>
                                        {testStatus === 'testing' ? 'sync' : testStatus === 'success' ? 'check_circle' : testStatus === 'error' ? 'cancel' : 'wifi_tethering'}
                                    </span>
                                    {t('settings.testConn')}
                                </button>
                            </div>
                            <p className="text-xs text-text-muted mt-1 flex items-center gap-1">
                                <span className="material-symbols-outlined text-[14px]">lock</span>
                                {t('settings.keyHint')}
                            </p>
                        </label>
                    </div>
                </section>

                {/* Export Section */}
                <section className="flex flex-col gap-4">
                    <div className="px-4 pb-2 pt-4 border-b border-border-light dark:border-border-dark flex items-center gap-3">
                        <span className="material-symbols-outlined text-primary">folder_open</span>
                        <h2 className="text-xl font-bold leading-tight tracking-tight">{t('settings.exportSection')}</h2>
                    </div>

                    <div className="px-4 py-2">
                        <div className="flex flex-col md:flex-row gap-3">
                            <div className="relative flex-1">
                                <input
                                    type="text"
                                    value={localConfig.output_path || ''}
                                    placeholder={defaultPath}
                                    onChange={(e) => setLocalConfig({ ...localConfig, output_path: e.target.value })}
                                    className="w-full rounded-lg border border-border-light dark:border-border-dark bg-slate-100 dark:bg-[#151c1e] text-text-muted px-4 h-14 text-base focus:outline-none font-mono text-sm placeholder:opacity-50"
                                />
                            </div>
                            <button
                                onClick={handleBrowse}
                                className="flex items-center justify-center gap-2 h-14 px-6 rounded-lg bg-primary/10 hover:bg-primary/20 text-primary border border-transparent hover:border-primary/30 transition-all font-medium whitespace-nowrap"
                            >
                                <span className="material-symbols-outlined">folder</span>
                                {t('settings.browse')}
                            </button>
                        </div>
                    </div>
                </section>

                {/* Hardware Section */}
                <section className="flex flex-col gap-4">
                    <div className="px-4 pb-2 pt-4 border-b border-border-light dark:border-border-dark flex items-center gap-3">
                        <span className="material-symbols-outlined text-primary">developer_board</span>
                        <h2 className="text-xl font-bold leading-tight tracking-tight">{t('settings.hardwareTarget')}</h2>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6 px-4 py-2">
                        <label className="flex flex-col flex-1 gap-2">
                            <p className="text-base font-medium leading-normal">{t('settings.hardwareTarget')}</p>
                            <div className="relative">
                                <select
                                    value={localConfig.hardware_target}
                                    onChange={(e) => setLocalConfig({ ...localConfig, hardware_target: e.target.value })}
                                    className="w-full appearance-none rounded-lg border border-border-light dark:border-border-dark bg-surface-light dark:bg-surface-dark px-4 h-14 text-base focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all cursor-pointer"
                                >
                                    <optgroup label="Multi-DSP (Path 1 & 2)">
                                        <option value="Helix Floor">Helix Floor / Rack</option>
                                        <option value="Helix LT">Helix LT</option>
                                    </optgroup>
                                    <optgroup label="Single-DSP (Path 1 Only)">
                                        <option value="HX Stomp">HX Stomp</option>
                                        <option value="HX Stomp XL">HX Stomp XL</option>
                                        <option value="HX Effects">HX Effects</option>
                                    </optgroup>
                                </select>
                                <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-text-muted">
                                    <span className="material-symbols-outlined">expand_more</span>
                                </div>
                            </div>
                            <p className="text-xs text-text-muted mt-1 flex items-center gap-1">
                                <span className="material-symbols-outlined text-[14px]">info</span>
                                {t('settings.hardwareHint')}
                            </p>
                        </label>

                        <label className="flex flex-col flex-1 gap-2">
                            <p className="text-base font-medium leading-normal">{t('settings.defaultExpPedal')}</p>
                            <div className="relative">
                                <select
                                    value={localConfig.default_exp_pedal}
                                    onChange={(e) => setLocalConfig({ ...localConfig, default_exp_pedal: parseInt(e.target.value) })}
                                    className="w-full appearance-none rounded-lg border border-border-light dark:border-border-dark bg-surface-light dark:bg-surface-dark px-4 h-14 text-base focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all cursor-pointer"
                                >
                                    <option value="0">{t('settings.expOptions.none')}</option>
                                    <option value="1">{t('settings.expOptions.exp1')}</option>
                                    <option value="2">{t('settings.expOptions.exp2')}</option>
                                    <option value="3">{t('settings.expOptions.exp3')}</option>
                                </select>
                                <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-text-muted">
                                    <span className="material-symbols-outlined">expand_more</span>
                                </div>
                            </div>
                            <p className="text-xs text-text-muted mt-1 flex items-center gap-1">
                                <span className="material-symbols-outlined text-[14px]">settings_input_component</span>
                                {t('settings.defaultExpHint')}
                            </p>
                        </label>
                    </div>
                </section>

                {/* Variax Section */}
                <section className="flex flex-col gap-4">
                    <div className="px-4 pb-2 pt-4 border-b border-border-light dark:border-border-dark flex items-center gap-3">
                        <div className="w-6 h-6 text-primary flex items-center justify-center">
                            <HelixIcons.Guitar className="w-full h-full" />
                        </div>
                        <h2 className="text-xl font-bold leading-tight tracking-tight">{t('settings.variaxSection')}</h2>
                    </div>

                    <div className="px-4 py-2 flex flex-col gap-6">
                        {/* Control Toggle */}
                        <div className="flex items-center justify-between p-4 bg-surface-light dark:bg-[#151c1e] rounded-xl border border-border-light dark:border-border-dark hover:border-primary/30 transition-all group">
                            <div className="flex flex-col gap-0.5">
                                <p className="text-base font-medium text-slate-900 dark:text-white group-hover:text-primary transition-colors">{t('settings.variaxEnabled')}</p>
                                <p className="text-xs text-text-muted">{t('settings.variaxEnabledHint')}</p>
                            </div>
                            <button
                                onClick={() => setLocalConfig({ ...localConfig, variax_enabled: !localConfig.variax_enabled })}
                                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 dark:focus:ring-offset-background-dark ${localConfig.variax_enabled ? 'bg-primary' : 'bg-slate-300 dark:bg-border-dark'}`}
                            >
                                <span
                                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${localConfig.variax_enabled ? 'translate-x-6' : 'translate-x-1'}`}
                                />
                            </button>
                        </div>

                        {/* Hardware Model selection */}
                        <label className={`flex flex-col gap-2 transition-opacity ${localConfig.variax_enabled ? 'opacity-100' : 'opacity-50 pointer-events-none'}`}>
                            <p className="text-base font-medium leading-normal">{t('settings.variaxHardwareModel')}</p>
                            <div className="relative">
                                <select
                                    value={localConfig.variax_hardware_model}
                                    onChange={(e) => setLocalConfig({ ...localConfig, variax_hardware_model: e.target.value })}
                                    className="w-full appearance-none rounded-lg border border-border-light dark:border-border-dark bg-surface-light dark:bg-surface-dark px-4 h-14 text-base focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-all cursor-pointer"
                                >
                                    <option value="Standard">Variax Standard</option>
                                    <option value="JTV">James Tyler Variax (JTV)</option>
                                    <option value="Shuriken">Variax Shuriken</option>
                                </select>
                                <div className="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-text-muted">
                                    <span className="material-symbols-outlined">expand_more</span>
                                </div>
                            </div>
                            <p className="text-xs text-text-muted mt-1 flex items-center gap-1">
                                <span className="material-symbols-outlined text-[14px]">info</span>
                                {t('settings.variaxHardwareHint')}
                            </p>
                        </label>
                    </div>
                </section>

                {/* Interface Section */}
                <section className="flex flex-col gap-4">
                    <div className="px-4 pb-2 pt-4 border-b border-border-light dark:border-border-dark flex items-center gap-3">
                        <span className="material-symbols-outlined text-primary">auto_fix_high</span>
                        <h2 className="text-xl font-bold leading-tight tracking-tight">{t('settings.interfaceSection')}</h2>
                    </div>

                    <div className="px-4 py-2">
                        <div className="flex flex-col flex-1 gap-4">
                            {/* Default Instrument Toggle */}
                            <div className="flex items-center justify-between p-4 bg-surface-light dark:bg-[#151c1e] rounded-xl border border-border-light dark:border-border-dark hover:border-primary/30 transition-all group">
                                <div className="flex flex-col gap-0.5">
                                    <p className="text-base font-medium text-slate-900 dark:text-white group-hover:text-primary transition-colors">{t('settings.defaultInstrument')}</p>
                                    <p className="text-xs text-text-muted">{t('settings.defaultInstrumentHint')}</p>
                                </div>
                                <div className="flex bg-slate-200 dark:bg-border-dark p-1 rounded-lg">
                                    <button
                                        onClick={() => setLocalConfig({ ...localConfig, default_instrument: "Guitar" })}
                                        className={`px-4 py-1.5 text-sm font-medium rounded-md transition-all ${localConfig.default_instrument === "Guitar" || !localConfig.default_instrument ? 'bg-white dark:bg-surface-dark shadow-sm text-primary' : 'text-text-muted hover:text-slate-900 dark:hover:text-white'}`}
                                    >
                                        Guitar
                                    </button>
                                    <button
                                        onClick={() => setLocalConfig({ ...localConfig, default_instrument: "Bass" })}
                                        className={`px-4 py-1.5 text-sm font-medium rounded-md transition-all ${localConfig.default_instrument === "Bass" ? 'bg-white dark:bg-surface-dark shadow-sm text-primary' : 'text-text-muted hover:text-slate-900 dark:hover:text-white'}`}
                                    >
                                        Bass
                                    </button>
                                </div>
                            </div>

                            {/* Lang Toggle */}
                            <div className="flex flex-col gap-2">
                                <p className="text-base font-medium leading-normal">Language</p>
                                <div className="flex gap-4">
                                    <button
                                        onClick={() => changeLang('en')}
                                        className={`px-4 py-2 rounded-lg border ${lang === 'en' ? 'bg-primary/20 border-primary text-primary' : 'border-border-dark'}`}
                                    >
                                        English
                                    </button>
                                    <button
                                        onClick={() => changeLang('fr')}
                                        className={`px-4 py-2 rounded-lg border ${lang === 'fr' ? 'bg-primary/20 border-primary text-primary' : 'border-border-dark'}`}
                                    >
                                        Français
                                    </button>
                                </div>
                            </div>

                            {/* Delete No Confirm Toggle */}
                            <div className="flex items-center justify-between p-4 bg-surface-light dark:bg-[#151c1e] rounded-xl border border-border-light dark:border-border-dark hover:border-primary/30 transition-all group">
                                <div className="flex flex-col gap-0.5">
                                    <p className="text-base font-medium text-slate-900 dark:text-white group-hover:text-primary transition-colors">{t('settings.deleteNoConfirm')}</p>
                                    <p className="text-xs text-text-muted">{t('settings.deleteNoConfirmHint')}</p>
                                </div>
                                <button
                                    onClick={() => setLocalConfig({ ...localConfig, delete_no_confirm: !localConfig.delete_no_confirm })}
                                    className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 dark:focus:ring-offset-background-dark ${localConfig.delete_no_confirm ? 'bg-primary' : 'bg-slate-300 dark:bg-border-dark'
                                        }`}
                                >
                                    <span
                                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${localConfig.delete_no_confirm ? 'translate-x-6' : 'translate-x-1'
                                            }`}
                                    />
                                </button>
                            </div>

                            {/* Incremental Save Toggle */}
                            <div className="flex items-center justify-between p-4 bg-surface-light dark:bg-[#151c1e] rounded-xl border border-border-light dark:border-border-dark hover:border-primary/30 transition-all group">
                                <div className="flex flex-col gap-0.5">
                                    <p className="text-base font-medium text-slate-900 dark:text-white group-hover:text-primary transition-colors">{t('settings.incrementalSave')}</p>
                                    <p className="text-xs text-text-muted">{t('settings.incrementalSaveHint')}</p>
                                </div>
                                <button
                                    onClick={() => setLocalConfig({ ...localConfig, incremental_save: !localConfig.incremental_save })}
                                    className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 dark:focus:ring-offset-background-dark ${localConfig.incremental_save ? 'bg-primary' : 'bg-slate-300 dark:bg-border-dark'
                                        }`}
                                >
                                    <span
                                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${localConfig.incremental_save ? 'translate-x-6' : 'translate-x-1'
                                            }`}
                                    />
                                </button>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Diagnostics Section */}
                <section className="flex flex-col gap-4">
                    <div className="px-4 pb-2 pt-4 border-b border-border-light dark:border-border-dark flex items-center gap-3">
                        <span className="material-symbols-outlined text-primary">bug_report</span>
                        <h2 className="text-xl font-bold leading-tight tracking-tight">{t('settings.diagnostics')}</h2>
                    </div>

                    <div className="px-4 py-2 flex flex-col gap-3">
                        <p className="text-sm text-text-muted">{t('settings.diagnosticsHint')}</p>
                        <div className="flex flex-wrap gap-3">
                            <button onClick={handleToggleLogs} className="flex items-center gap-2 h-10 px-4 rounded-lg bg-primary/10 hover:bg-primary/20 text-primary border border-primary/20 transition-all font-medium text-sm">
                                <span className="material-symbols-outlined">{showLogs ? 'visibility_off' : 'visibility'}</span>
                                {t(showLogs ? 'settings.hideLogs' : 'settings.showLogs')}
                            </button>
                            {logPath && <button onClick={() => GxOpenFolderOfFile(logPath)} className="flex items-center gap-2 h-10 px-4 rounded-lg border border-border-light dark:border-border-dark hover:border-primary hover:text-primary transition-all font-medium text-sm">
                                <span className="material-symbols-outlined">folder_open</span>
                                {t('settings.openLogFolder')}
                            </button>}
                        </div>
                        {showLogs && <div className="flex flex-col gap-3">
                            <p className="text-xs text-text-muted font-mono break-all">{t('settings.logPath')}: {logPath}</p>
                            <pre className="min-h-32 max-h-72 overflow-auto rounded-lg border border-border-light dark:border-border-dark bg-slate-100 dark:bg-[#111719] p-4 text-xs text-slate-700 dark:text-slate-300 whitespace-pre-wrap font-mono">{loadingLogs ? '...' : logContent || t('settings.noLogs')}</pre>
                            <p className="text-xs text-amber-500">{t('settings.logWarning')}</p>
                            <div className="flex flex-wrap gap-3">
                                <button onClick={refreshLogs} disabled={loadingLogs} className="flex items-center gap-2 h-9 px-3 rounded-lg border border-border-light dark:border-border-dark hover:border-primary hover:text-primary transition-all text-sm disabled:opacity-50">
                                    <span className={`material-symbols-outlined text-[18px] ${loadingLogs ? 'animate-spin' : ''}`}>sync</span>
                                    {t('settings.refreshLogs')}
                                </button>
                                <button onClick={handleCopyLogs} disabled={!logContent} className="flex items-center gap-2 h-9 px-3 rounded-lg border border-border-light dark:border-border-dark hover:border-primary hover:text-primary transition-all text-sm disabled:opacity-50">
                                    <span className="material-symbols-outlined text-[18px]">content_copy</span>
                                    {logsCopied ? t('settings.logsCopied') : t('settings.copyLogs')}
                                </button>
                                <button onClick={handleClearLogs} className="flex items-center gap-2 h-9 px-3 rounded-lg border border-red-500/40 text-red-400 hover:bg-red-500/10 transition-all text-sm">
                                    <span className="material-symbols-outlined text-[18px]">delete</span>
                                    {t('settings.clearLogs')}
                                </button>
                            </div>
                        </div>}
                    </div>
                </section>

                {/* Footer Actions */}
                <div className="sticky bottom-4 mx-4 mt-6 p-4 rounded-xl bg-surface-light dark:bg-[#1a2325] border border-border-light dark:border-border-dark shadow-xl flex items-center justify-end gap-4 backdrop-blur-md bg-opacity-90 dark:bg-opacity-90 z-40">
                    <div className="flex w-full sm:w-auto gap-3">
                        <button
                            className="flex-1 sm:flex-none h-10 px-6 rounded-lg border border-border-light dark:border-border-dark text-slate-700 dark:text-white hover:bg-slate-100 dark:hover:bg-white/5 transition-colors font-medium text-sm"
                            onClick={() => window.location.reload()} // Quick way to cancel
                        >
                            {t('settings.cancel')}
                        </button>
                        <button
                            disabled={saving}
                            onClick={handleSave}
                            className="flex-1 sm:flex-none h-10 px-6 rounded-lg bg-primary text-black hover:bg-cyan-300 transition-colors font-bold text-sm shadow-[0_0_15px_rgba(19,200,236,0.3)] hover:shadow-[0_0_20px_rgba(19,200,236,0.5)]"
                        >
                            {saving ? '...' : t('settings.save')}
                        </button>
                    </div>
                </div>
            </div>
        </main>
    );
};

export default Settings;
