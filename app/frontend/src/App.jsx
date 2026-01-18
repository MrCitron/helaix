import { useState, useEffect } from 'react';
import { GxGetConfig } from '../wailsjs/go/main/App';
import MainScreen from './components/MainScreen';
import Settings from './components/Settings';
import Sidebar from './components/Sidebar';
import DeleteConfirmationModal from './components/DeleteConfirmationModal';

function App() {
    const [config, setConfig] = useState(null);
    const [view, setView] = useState('chat'); // 'chat' | 'settings'
    const [loading, setLoading] = useState(true);

    // Chat History State
    const [chats, setChats] = useState(() => {
        const saved = localStorage.getItem('helAIx_chats');
        return saved ? JSON.parse(saved) : [];
    });
    const [currentChatId, setCurrentChatId] = useState(() => {
        const saved = localStorage.getItem('helAIx_currentChatId');
        return saved || null;
    });

    const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(window.innerWidth < 1024);
    const [deleteConfirmId, setDeleteConfirmId] = useState(null);

    // Save history whenever it changes
    useEffect(() => {
        localStorage.setItem('helAIx_chats', JSON.stringify(chats));
    }, [chats]);

    useEffect(() => {
        if (currentChatId) {
            localStorage.setItem('helAIx_currentChatId', currentChatId);
        } else {
            localStorage.removeItem('helAIx_currentChatId');
        }
    }, [currentChatId]);

    useEffect(() => {
        async function loadConfig() {
            try {
                const cfg = await GxGetConfig();
                if (!cfg || !cfg.api_key) {
                    setView('settings');
                }
                setConfig(cfg);
            } catch (err) {
                console.error("Failed to load config", err);
            } finally {
                setLoading(false);
            }
        }
        loadConfig();
    }, []);

    // Always use dark mode
    useEffect(() => {
        document.documentElement.classList.add('dark');
    }, []);

    // Auto-collapse sidebar on resize
    useEffect(() => {
        const handleResize = () => {
            if (window.innerWidth < 1024) {
                setIsSidebarCollapsed(true);
            }
        };
        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, []);

    const handleConfigSave = (newConfig) => {
        setConfig(newConfig);
        if (view === 'settings' && newConfig.api_key) {
            setView('chat');
        }
    };

    // Chat Handlers
    const createNewChat = () => {
        const newId = Date.now().toString();
        const newChat = {
            id: newId,
            name: "New Chat",
            messages: [],
            createdAt: new Date().toISOString()
        };
        setChats(prev => [newChat, ...prev]);
        setCurrentChatId(newId);
        setView('chat');
    };

    const switchChat = (id) => {
        setCurrentChatId(id);
        setView('chat');
    };

    const updateCurrentChat = (updateFn) => {
        setChats(prev => prev.map(chat => {
            if (chat.id === currentChatId) {
                return updateFn(chat);
            }
            return chat;
        }));
    };

    const handleDeleteChat = (e, id) => {
        e.stopPropagation();
        if (config?.delete_no_confirm) {
            confirmDeleteChat(id);
        } else {
            setDeleteConfirmId(id);
        }
    };

    const confirmDeleteChat = (id) => {
        const chatToDelete = id || deleteConfirmId;
        if (!chatToDelete) return;

        setChats(prev => {
            const updated = prev.filter(c => c.id !== chatToDelete);

            // If we deleted the current chat, switch to another one or create new
            if (chatToDelete === currentChatId) {
                if (updated.length > 0) {
                    setCurrentChatId(updated[0].id);
                } else {
                    setCurrentChatId(null);
                }
            }
            return updated;
        });

        setDeleteConfirmId(null);
    };

    const currentChat = chats.find(c => c.id === currentChatId) || null;

    if (loading) return (
        <div className="bg-background-light dark:bg-background-dark min-h-screen text-slate-900 dark:text-white flex items-center justify-center font-display">
            <span className="material-symbols-outlined animate-spin text-primary mr-2">sync</span>
            Loading...
        </div>
    );

    return (
        <div className="flex h-screen bg-background-light dark:bg-background-dark text-slate-900 dark:text-white font-display overflow-hidden">
            <Sidebar
                currentView={view}
                onViewChange={setView}
                recentChats={chats}
                currentChatId={currentChatId}
                onNewChat={createNewChat}
                onSwitchChat={switchChat}
                onDeleteChat={handleDeleteChat}
                isCollapsed={isSidebarCollapsed}
                onToggleCollapse={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
            />

            <main className="flex-1 flex flex-col h-full overflow-hidden">
                {view === 'chat' && (
                    <MainScreen
                        config={config}
                        chatData={currentChat}
                        onUpdateChat={updateCurrentChat}
                        onNewChat={createNewChat}
                    />
                )}
                {view === 'settings' && <Settings config={config || {}} onSave={handleConfigSave} />}
            </main>

            <DeleteConfirmationModal
                isOpen={!!deleteConfirmId}
                onClose={() => setDeleteConfirmId(null)}
                onConfirm={() => confirmDeleteChat()}
                chatName={chats.find(c => c.id === deleteConfirmId)?.name}
            />
        </div>
    );
}

export default App;
