import React from 'react';
import { useI18n } from '../i18n';

const DeleteConfirmationModal = ({ isOpen, onClose, onConfirm, chatName }) => {
    const { t } = useI18n();

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="bg-[#111718] border border-[#283639] w-full max-w-md rounded-2xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200">
                <div className="p-6">
                    <div className="flex items-center gap-3 mb-4">
                        <div className="size-10 rounded-full bg-red-500/10 flex items-center justify-center text-red-500">
                            <span className="material-symbols-outlined">delete_forever</span>
                        </div>
                        <h3 className="text-xl font-bold text-white">{t('chat.deleteConfirmTitle')}</h3>
                    </div>

                    <p className="text-[#9db4b9] text-base leading-relaxed mb-6">
                        {t('chat.deleteConfirmMsg')}
                        {chatName && (
                            <span className="block mt-2 font-bold text-white italic">"{chatName}"</span>
                        )}
                    </p>

                    <div className="flex gap-3 justify-end">
                        <button
                            onClick={onClose}
                            className="px-6 py-2.5 rounded-xl border border-[#283639] text-[#9db4b9] hover:bg-[#1d2628] hover:text-white transition-all font-medium"
                        >
                            {t('chat.deleteConfirmCancel')}
                        </button>
                        <button
                            onClick={onConfirm}
                            className="px-6 py-2.5 rounded-xl bg-red-500 hover:bg-red-600 text-white transition-all font-bold shadow-lg shadow-red-500/20"
                        >
                            {t('chat.deleteConfirmOk')}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default DeleteConfirmationModal;
