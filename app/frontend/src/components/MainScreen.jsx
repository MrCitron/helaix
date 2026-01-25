import React from 'react';
import { GxChatSoundEngineer, GxChatPresetEngineer, GxSaveFile } from '../../wailsjs/go/main/App';
import { useI18n } from '../i18n';
import DesignVisualizer from './DesignVisualizer';
import MessageVisualizer from './MessageVisualizer';
import ChatInput from './ChatInput';
import ExportModal from './ExportModal';

const MainScreen = ({ config, chatData, onUpdateChat, onNewChat }) => {
    const { t } = useI18n();
    const [loading, setLoading] = React.useState(false);
    const [showExportModal, setShowExportModal] = React.useState(false);
    const [lastExportPath, setLastExportPath] = React.useState('');
    const bottomRef = React.useRef(null);

    const messages = chatData?.messages || [];
    const stage = chatData?.stage || 'design'; // 'design' or 'build'

    React.useEffect(() => {
        if (chatData && messages.length === 0) {
            onUpdateChat(chat => ({
                ...chat,
                stage: 'design',
                messages: [{
                    id: 'welcome',
                    role: 'assistant',
                    agent: 'sound_engineer',
                    content: t('chat.welcome'),
                    hint: t('chat.welcomeHint')
                }]
            }));
        }
    }, [chatData?.id, messages.length, onUpdateChat, t]);

    React.useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages, loading]);

    const formatHistory = (msgs) => {
        return msgs.map(m => ({
            role: m.role,
            content: m.content || ""
        })).filter(m => m.content !== "");
    };

    const handleSendMessage = async (text) => {
        if (!text || !chatData) return;

        const isFirstUserMsg = messages.filter(m => m.role === 'user').length === 0;
        if (isFirstUserMsg) {
            const shortName = text.substring(0, 30).trim() + (text.length > 30 ? "..." : "");
            onUpdateChat(chat => ({ ...chat, name: shortName }));
        }

        const userMsg = { id: Date.now(), role: 'user', content: text };
        const updatedMessages = [...messages, userMsg];

        onUpdateChat(chat => ({
            ...chat,
            messages: updatedMessages
        }));

        setLoading(true);

        try {
            if (stage === 'design') {
                const design = await GxChatSoundEngineer(formatHistory(updatedMessages));
                const aiMsg = {
                    id: Date.now() + 1,
                    role: 'assistant',
                    agent: 'sound_engineer',
                    design: design,
                    content: design.explanation
                };
                onUpdateChat(chat => ({
                    ...chat,
                    messages: [...updatedMessages, aiMsg]
                }));
            } else {
                const latestDesign = [...messages].reverse().find(m => m.design)?.design;
                const presetName = latestDesign?.suggested_name || "HelAIx Preset";
                const preset = await GxChatPresetEngineer(latestDesign, presetName, formatHistory(updatedMessages));
                const aiMsg = {
                    id: Date.now() + 1,
                    role: 'assistant',
                    agent: 'preset_engineer',
                    preset: preset,
                    content: "I've refined the technical preset based on your feedback."
                };
                onUpdateChat(chat => ({
                    ...chat,
                    messages: [...updatedMessages, aiMsg]
                }));
            }
        } catch (err) {
            onUpdateChat(chat => ({
                ...chat,
                messages: [...updatedMessages, {
                    id: Date.now() + 2,
                    role: 'assistant',
                    error: "AI failed: " + err
                }]
            }));
        } finally {
            setLoading(false);
        }
    };

    const handleBuildPreset = async (messageId, design) => {
        setLoading(true);
        try {
            const presetName = design.suggested_name || "HelAIx Preset";
            const preset = await GxChatPresetEngineer(design, presetName, []);

            onUpdateChat(chat => ({
                ...chat,
                stage: 'build',
                messages: chat.messages.map(m =>
                    m.id === messageId ? { ...m, preset: preset } : m
                )
            }));
        } catch (err) {
            alert("Build failed: " + err);
        } finally {
            setLoading(false);
        }
    };

    const handleExportHlx = async (preset) => {
        setLoading(true);
        try {
            const safeName = (preset.data?.meta?.name || "HelAIx_Preset").replace(/\s+/g, '_');
            const filename = safeName + ".hlx";
            const filePath = await GxSaveFile(preset, filename);
            setLastExportPath(filePath);
            setShowExportModal(true);
        } catch (err) {
            alert("Export failed: " + err);
        } finally {
            setLoading(false);
        }
    };

    if (!chatData) {
        return (
            <div className="flex-1 flex flex-col items-center justify-center p-8 bg-background-light dark:bg-background-dark text-center">
                <div className="size-16 rounded-3xl bg-primary/10 flex items-center justify-center text-primary mb-6">
                    <span className="material-symbols-outlined text-4xl">graphic_eq</span>
                </div>
                <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">{t('chat.welcome')}</h2>
                <p className="text-slate-600 dark:text-text-muted mb-8 max-w-md">{t('chat.welcomeHint')}</p>
                <button
                    onClick={onNewChat}
                    className="flex items-center gap-2 px-6 py-3 bg-primary hover:bg-[#0fb3d4] text-[#111718] font-bold rounded-xl transition-all shadow-lg shadow-primary/20"
                >
                    <span className="material-symbols-outlined">add</span>
                    {t('nav.newChat')}
                </button>
            </div>
        );
    }

    return (
        <div className="flex-1 flex flex-col h-full bg-background-light dark:bg-background-dark overflow-hidden relative">
            <header className="h-16 shrink-0 border-b border-slate-300 dark:border-border-dark flex items-center justify-between px-6 bg-white/90 dark:bg-background-dark/90 backdrop-blur-sm z-10">
                <div className="flex items-center gap-3">
                    <div className="size-8 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
                        <span className="material-symbols-outlined text-lg">
                            smart_toy
                        </span>
                    </div>
                    <div className="min-w-0">
                        <h2 className="text-slate-900 dark:text-white text-base font-bold leading-tight truncate">
                            {chatData.name || t('chat.currentChat')}
                        </h2>
                        <div className="flex items-center gap-2">
                            <span className="text-primary text-[10px] font-bold uppercase tracking-widest">
                                {t('chat.aiName')}
                            </span>
                            <span className="text-slate-400 dark:text-[#2b3d41] text-[10px]">•</span>
                            <p className="text-slate-600 dark:text-text-muted text-[10px]">Active Discussion</p>
                        </div>
                    </div>
                </div>
            </header>

            <div className="flex-1 flex overflow-hidden">
                <div className="flex-1 flex flex-col overflow-hidden border-r border-slate-300 dark:border-border-dark">
                    <div className="flex-1 overflow-y-auto overflow-x-hidden p-6 space-y-8 pb-32">
                        {messages.map((msg) => (
                            <div key={msg.id} className={`flex gap-4 ${msg.role === 'assistant' ? 'justify-start' : 'justify-end'}`}>
                                {msg.role === 'assistant' && (
                                    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-indigo-600 flex items-center justify-center flex-shrink-0 border border-slate-300 dark:border-border-dark">
                                        <span className="material-symbols-outlined text-white text-xl">smart_toy</span>
                                    </div>
                                )}

                                <div className={`flex flex-col gap-1 max-w-[85%] ${msg.role === 'user' ? 'items-end' : 'items-start'}`}>
                                    <span className="text-slate-500 dark:text-text-muted text-xs font-medium px-1">
                                        {msg.role === 'assistant'
                                            ? t('chat.aiName')
                                            : t('chat.userName')
                                        }
                                    </span>
                                    <div className={`p-4 rounded-2xl shadow-sm text-[15px] leading-relaxed font-body break-words overflow-hidden ${msg.role === 'assistant'
                                        ? 'bg-slate-100 dark:bg-border-dark text-slate-900 dark:text-white/90 rounded-bl-none'
                                        : 'bg-primary/10 border border-primary/20 text-slate-900 dark:text-white rounded-br-none'
                                        }`}>
                                        {msg.content && <p className="whitespace-pre-wrap break-words">{msg.content}</p>}
                                        {msg.hint && <p className="mt-2 text-slate-500 dark:text-text-secondary text-xs italic">{msg.hint}</p>}

                                        {(msg.design || msg.preset) && (
                                            <div className="mt-4 pt-4 border-t border-slate-300 dark:border-indigo-800/50">
                                                <MessageVisualizer
                                                    msg={msg}
                                                    messages={messages}
                                                    onBuildPreset={handleBuildPreset}
                                                    onExportHlx={handleExportHlx}
                                                />
                                            </div>
                                        )}

                                        {msg.error && (
                                            <div className="text-red-400 bg-red-900/20 p-2 rounded border border-red-900/50 text-sm">
                                                {msg.error}
                                            </div>
                                        )}
                                    </div>
                                </div>

                                {msg.role === 'user' && (
                                    <div className="w-10 h-10 rounded-full bg-slate-200 dark:bg-border-dark flex items-center justify-center flex-shrink-0 border border-slate-300 dark:border-[#4a5e63]">
                                        <span className="material-symbols-outlined text-slate-600 dark:text-text-secondary text-xl">person</span>
                                    </div>
                                )}
                            </div>
                        ))}
                        {loading && (
                            <div className="flex gap-4 justify-start">
                                <div className="w-10 h-10 rounded-full bg-slate-100 dark:bg-surface-dark flex items-center justify-center border border-slate-300 dark:border-border-dark">
                                    <span className="material-symbols-outlined text-primary text-xl animate-spin">sync</span>
                                </div>
                                <div className="bg-slate-100 dark:bg-border-dark p-4 rounded-2xl rounded-bl-none flex items-center gap-2">
                                    <span className="w-2 h-2 bg-primary rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></span>
                                    <span className="w-2 h-2 bg-primary rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></span>
                                    <span className="w-2 h-2 bg-primary rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></span>
                                </div>
                            </div>
                        )}
                        <div ref={bottomRef} />
                    </div>

                    <div className="absolute bottom-0 left-0 right-0 p-6 pt-2 bg-gradient-to-t from-background-light dark:from-background-dark via-background-light dark:via-background-dark to-transparent z-50">
                        <div className="max-w-3xl mx-auto flex flex-col gap-2">
                            <ChatInput onSend={handleSendMessage} loading={loading} />
                            <p className="text-center text-slate-500 dark:text-text-muted text-[10px] mt-1 font-body">
                                {t('chat.errors.ia')}
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <ExportModal
                isOpen={showExportModal}
                onClose={() => setShowExportModal(false)}
                filePath={lastExportPath}
            />
        </div>
    );
};

export default MainScreen;
