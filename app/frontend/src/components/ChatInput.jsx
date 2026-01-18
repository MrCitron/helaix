import React, { useState, useRef } from 'react';
import { useI18n } from '../i18n';

const ChatInput = ({ onSend, loading }) => {
    const { t } = useI18n();
    const [text, setText] = useState('');
    const textareaRef = useRef(null);

    const handleSend = () => {
        if (text.trim() && !loading) {
            onSend(text.trim());
            setText('');
            if (textareaRef.current) {
                textareaRef.current.style.height = '56px';
            }
        }
    };

    const handleKeyDown = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    };

    const handleChange = (e) => {
        setText(e.target.value);
        // Auto-resize
        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
            textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`;
        }
    };

    return (
        <div className="relative bg-white dark:bg-surface-dark rounded-xl border border-slate-300 dark:border-[#3f5256] focus-within:border-primary/50 focus-within:ring-1 focus-within:ring-primary/50 transition-all shadow-lg overflow-hidden">
            <textarea
                ref={textareaRef}
                value={text}
                onChange={handleChange}
                onKeyDown={handleKeyDown}
                className="w-full bg-transparent text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-text-muted text-sm p-4 pr-12 rounded-xl focus:outline-none resize-none font-body transition-all"
                placeholder={t('chat.placeholder')}
                rows="1"
                style={{ minHeight: '56px' }}
            />
            <button
                onClick={handleSend}
                disabled={!text.trim() || loading}
                className={`absolute right-2 bottom-2 p-2 rounded-lg transition-all flex items-center justify-center ${text.trim() && !loading
                        ? 'bg-primary text-slate-900 hover:bg-[#0fb3d4] shadow-lg shadow-primary/20'
                        : 'bg-slate-200 dark:bg-border-dark text-slate-400 dark:text-text-muted cursor-not-allowed'
                    }`}
            >
                <span className="material-symbols-outlined text-[20px]">
                    {loading ? 'sync' : 'arrow_upward'}
                </span>
            </button>
        </div>
    );
};

export default ChatInput;
