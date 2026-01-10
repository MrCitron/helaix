import React from 'react';
import { useI18n } from '../i18n';

const Sidebar = ({ currentView, onViewChange, recentChats = [], currentChatId, onNewChat, onSwitchChat, onDeleteChat, isCollapsed, onToggleCollapse }) => {
    const { t } = useI18n();

    const navItems = [
        { id: 'settings', label: t('nav.settings'), icon: 'settings' },
    ];

    return (
        <aside className={`${isCollapsed ? 'w-20' : 'w-64'} h-full bg-surface-light dark:bg-background-dark border-r border-slate-300 dark:border-border-dark flex flex-col justify-between p-4 shrink-0 z-20 transition-all duration-300 relative overflow-visible`}>
            {/* Collapse Toggle */}
            <button
                onClick={onToggleCollapse}
                className="absolute -right-3 top-20 size-6 bg-slate-200 dark:bg-border-dark border border-slate-300 dark:border-[#3b4f54] rounded-full flex items-center justify-center text-primary hover:text-white transition-all z-30 shadow-md"
            >
                <span className="material-symbols-outlined text-[16px]">
                    {isCollapsed ? 'chevron_right' : 'chevron_left'}
                </span>
            </button>

            <div className="flex flex-col gap-8">
                {/* App Header */}
                <div className={`flex flex-col gap-1 px-2 ${isCollapsed ? 'items-center' : ''}`}>
                    <div className="flex items-center gap-2 text-primary cursor-pointer" onClick={() => onViewChange('chat')}>
                        <span className="material-symbols-outlined text-3xl">graphic_eq</span>
                        {!isCollapsed && <h1 className="text-slate-900 dark:text-white text-xl font-bold tracking-tight">{t('appName')}</h1>}
                    </div>
                    {!isCollapsed && <p className="text-slate-600 dark:text-text-secondary text-xs font-normal leading-normal">{t('tagline')}</p>}
                </div>

                {/* Primary Action */}
                <button
                    onClick={onNewChat}
                    title={isCollapsed ? t('nav.newChat') : ''}
                    className={`flex items-center gap-3 rounded-xl bg-primary hover:bg-[#0fb3d4] text-slate-900 transition-all font-bold shadow-lg shadow-primary/10 ${isCollapsed ? 'p-3 justify-center' : 'px-3 py-2.5'}`}
                >
                    <span className="material-symbols-outlined">add_circle</span>
                    {!isCollapsed && <p className="text-sm">{t('nav.newChat')}</p>}
                </button>

                {/* Navigation */}
                <nav className="flex flex-col gap-2">
                    {navItems.map((item) => (
                        <button
                            key={item.id}
                            onClick={() => onViewChange(item.id)}
                            title={isCollapsed ? item.label : ''}
                            className={`flex items-center gap-3 rounded-lg transition-colors group ${isCollapsed ? 'p-3 justify-center' : 'px-3 py-2.5'} ${currentView === item.id
                                ? 'bg-slate-200 dark:bg-border-dark text-slate-900 dark:text-white'
                                : 'text-slate-600 dark:text-text-secondary hover:bg-slate-100 dark:hover:bg-surface-dark hover:text-slate-900 dark:hover:text-white'
                                }`}
                        >
                            <span className={`material-symbols-outlined transition-colors ${currentView === item.id ? 'text-primary' : 'group-hover:text-white'
                                }`}>
                                {item.icon}
                            </span>
                            {!isCollapsed && <p className="text-sm font-medium">{item.label}</p>}
                        </button>
                    ))}
                </nav>

                {/* History Section */}
                <div className={`flex flex-col gap-2 px-1 mt-2 ${isCollapsed ? 'items-center' : ''}`}>
                    {!isCollapsed && <p className="text-slate-500 dark:text-text-muted text-[10px] font-bold uppercase tracking-widest px-2 mb-1">{t('nav.recent')}</p>}
                    <div className={`flex flex-col gap-1 overflow-y-auto max-h-[400px] scrollbar-hide w-full ${isCollapsed ? 'items-center overflow-x-hidden' : ''}`}>
                        {recentChats.map((chat) => (
                            <div
                                key={chat.id}
                                className={`group/item relative flex items-center justify-between transition-all cursor-pointer ${isCollapsed
                                    ? 'p-3 justify-center rounded-none'
                                    : 'px-3 py-2 rounded-lg'
                                    } ${currentChatId === chat.id && currentView === 'chat'
                                        ? (isCollapsed ? 'text-primary border-r-2 border-primary' : 'bg-slate-200 dark:bg-border-dark/50 text-slate-900 dark:text-white border-l-2 border-primary pl-2')
                                        : 'text-slate-600 dark:text-text-secondary hover:bg-slate-100 dark:hover:bg-surface-dark hover:text-slate-900 dark:hover:text-white'
                                    }`}
                                onClick={() => onSwitchChat(chat.id)}
                                title={isCollapsed ? chat.name : ''}
                            >
                                {!isCollapsed && <span className="truncate flex-1">{chat.name}</span>}
                                {isCollapsed ? (
                                    <span className={`material-symbols-outlined text-[18px] ${currentChatId === chat.id && currentView === 'chat' ? 'text-primary' : ''}`}>chat</span>
                                ) : (
                                    <button
                                        onClick={(e) => onDeleteChat(e, chat.id)}
                                        className="opacity-0 group-hover/item:opacity-100 p-1 hover:text-red-500 transition-all rounded-md hover:bg-red-500/10"
                                        title={t('chat.deleteChat')}
                                    >
                                        <span className="material-symbols-outlined text-[16px]">delete</span>
                                    </button>
                                )}
                            </div>
                        ))}
                        {recentChats.length === 0 && !isCollapsed && (
                            <p className="px-3 py-4 text-[11px] text-slate-500 dark:text-text-muted italic text-center border border-dashed border-slate-300 dark:border-border-dark rounded-xl mr-2">
                                No history yet
                            </p>
                        )}
                        {recentChats.length === 0 && isCollapsed && (
                            <span className="material-symbols-outlined text-slate-300 dark:text-border-dark">history</span>
                        )}
                    </div>
                </div>
            </div>
        </aside>
    );
};

export default Sidebar;
