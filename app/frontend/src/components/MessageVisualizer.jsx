import React, { useState } from 'react';
import DesignVisualizer from './DesignVisualizer';
import PresetVisualizer from './PresetVisualizer';
import { useI18n } from '../i18n';

const MessageVisualizer = ({ msg, messages, onBuildPreset, onExportHlx }) => {
    const { t } = useI18n();
    const [activeSnapIdx, setActiveSnapIdx] = useState(0);

    // Resolve snapshots: use msg.design if available, otherwise find the latest design in history
    const design = msg.design || messages.findLast(m => m.design && (m.id < msg.id || m.id === msg.id))?.design;
    const snapshots = design?.snapshots || [];
    const activeSnapshot = snapshots.length > 0 ? { ...snapshots[activeSnapIdx], index: activeSnapIdx } : null;

    return (
        <div className="flex flex-col gap-4 w-full">
            {/* Unified Snapshot Selector */}
            {snapshots.length > 0 && (
                <div className="flex flex-wrap gap-2 px-1">
                    {snapshots.map((snap, idx) => (
                        <button
                            key={idx}
                            onClick={() => setActiveSnapIdx(idx)}
                            className={`px-3 py-1.5 rounded-lg text-[11px] font-bold transition-all border
                                ${activeSnapIdx === idx
                                    ? 'bg-primary text-slate-900 border-primary shadow-[0_0_10px_rgba(19,200,236,0.3)]'
                                    : 'bg-slate-100 dark:bg-surface-dark text-slate-500 dark:text-text-muted border-slate-300 dark:border-border-dark hover:border-primary/40'
                                }`}
                        >
                            {snap.name.toUpperCase()}
                        </button>
                    ))}
                </div>
            )}

            {/* Design View (Real Chain) */}
            {msg.design && (
                <div className="w-full">
                    <p className="mb-3 text-sm text-slate-500 dark:text-text-secondary">{t('chat.proposedChain')}</p>
                    <DesignVisualizer
                        design={msg.design}
                        activeSnapIdx={activeSnapIdx}
                        onGenerate={msg.preset ? null : () => onBuildPreset(msg.id, msg.design)}
                        hideSelector={true} // New prop to hide internal selector
                    />
                </div>
            )}

            {/* Build View (Helix Chain) */}
            {msg.preset && (
                <div className="space-y-4 w-full pt-4 border-t border-slate-300 dark:border-indigo-800/50">
                    <PresetVisualizer
                        preset={msg.preset}
                        design={msg.design || messages.findLast(m => m.design && m.id < msg.id)?.design}
                        activeSnapIdx={activeSnapIdx}
                        compact={true}
                        hideSelector={true} // New prop to hide internal selector
                    />
                    <button
                        onClick={() => onExportHlx(msg.preset)}
                        className="w-full bg-primary hover:bg-[#0fb3d4] text-slate-900 h-10 px-4 rounded-lg flex items-center justify-center gap-2 font-bold transition-all shadow-lg shadow-primary/10"
                    >
                        <span className="material-symbols-outlined text-[18px]">download</span>
                        {t('chat.exportBtn')}
                    </button>
                </div>
            )}
        </div>
    );
};

export default MessageVisualizer;
