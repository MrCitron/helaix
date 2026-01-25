import React, { useState } from 'react';
import { HelixIcons } from './IconLibrary';
import { useI18n } from '../i18n';

const typeConfig = {
    dynamics: { Icon: HelixIcons.Dynamics, color: 'bg-[#FFC107]', hoverColor: 'hover:bg-[#FFD54F]', borderColor: 'border-[#FFA000]', text: 'text-[#FFC107]', label: 'DYN' },
    pedal: { Icon: HelixIcons.Distortion, color: 'bg-[#FF5722]', hoverColor: 'hover:bg-[#FF7043]', borderColor: 'border-[#E64A19]', text: 'text-[#FF7043]', label: 'DIST' },
    amp: { Icon: HelixIcons.Amp, color: 'bg-[#FF5252]', hoverColor: 'hover:bg-[#FF8A80]', borderColor: 'border-[#D32F2F]', text: 'text-[#FF8A80]', label: 'AMP' },
    delay: { Icon: HelixIcons.Delay, color: 'bg-[#4CAF50]', hoverColor: 'hover:bg-[#81C784]', borderColor: 'border-[#388E3C]', text: 'text-[#81C784]', label: 'DLY' },
    reverb: { Icon: HelixIcons.Reverb, color: 'bg-[#FF9800]', hoverColor: 'hover:bg-[#FFB74D]', borderColor: 'border-[#F57C00]', text: 'text-[#FFB74D]', label: 'REV' },
    modulation: { Icon: HelixIcons.Modulation, color: 'bg-[#2196F3]', hoverColor: 'hover:bg-[#64B5F6]', borderColor: 'border-[#1976D2]', text: 'text-[#64B5F6]', label: 'MOD' },
    cab: { Icon: HelixIcons.Cab, color: 'bg-[#D32F2F]', hoverColor: 'hover:bg-[#EF5350]', borderColor: 'border-[#B71C1C]', text: 'text-[#EF5350]', label: 'CAB' },
    variax: { Icon: HelixIcons.Guitar, color: 'bg-[#9C27B0]', hoverColor: 'hover:bg-[#AB47BC]', borderColor: 'border-[#7B1FA2]', text: 'text-[#AB47BC]', label: 'VARX' },
    default: { Icon: HelixIcons.FX, color: 'bg-[#607D8B]', hoverColor: 'hover:bg-[#78909C]', borderColor: 'border-[#455A64]', text: 'text-[#B0BEC5]', label: 'FX' }
};

const DesignVisualizer = ({ design, onGenerate, activeSnapIdx: propsActiveSnapIdx, hideSelector = false }) => {
    const { t } = useI18n();
    const [localActiveSnapIdx, setLocalActiveSnapIdx] = useState(0);

    const activeSnapIdx = propsActiveSnapIdx !== undefined ? propsActiveSnapIdx : localActiveSnapIdx;
    const setActiveSnapIdx = propsActiveSnapIdx !== undefined ? () => { } : setLocalActiveSnapIdx;

    if (!design || !design.chain) return null;

    const snapshots = design.snapshots || [];
    const activeSnapshot = snapshots.length > 0 ? snapshots[activeSnapIdx] : null;

    return (
        <div className="flex flex-col gap-6 w-full">
            {/* Snapshot Selector */}
            {snapshots.length > 0 && !hideSelector && (
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

            <div className="relative w-full rounded-xl border border-slate-300 dark:border-[#3f5256] bg-white dark:bg-background-dark overflow-hidden">
                <div className="flex justify-between items-center px-6 pt-6 mb-2 opacity-80">
                    <div className="flex flex-col gap-1">
                        <h3 className="text-slate-500 dark:text-text-muted uppercase text-[10px] font-bold tracking-[0.2em]">{t('chat.realChain')}</h3>
                        {activeSnapshot && (
                            <span className="text-primary text-[10px] font-bold bg-primary/10 px-1.5 py-0.5 rounded w-fit">
                                {activeSnapshot.name} Mode
                            </span>
                        )}
                    </div>
                    {/* Guitar & Tuning Info */}
                    {(() => {
                        const currentModel = activeSnapshot?.guitar_model || design.guitar_model;
                        const currentTuning = activeSnapshot?.tuning || design.tuning;

                        if (!currentModel && !currentTuning) return null;

                        return (
                            <div className="flex items-center gap-3 text-primary text-[11px] font-bold">
                                {currentModel && (
                                    <div className="flex items-center gap-1.5 bg-primary/10 px-2 py-1 rounded">
                                        <span>{currentModel}</span>
                                    </div>
                                )}
                                <div className="flex items-center gap-1">
                                    <span className="material-symbols-outlined text-sm">tune</span>
                                    <span>{currentTuning && currentTuning !== 'Standard' ? currentTuning : 'Standard'}</span>
                                </div>
                            </div>
                        );
                    })()}
                </div>

                <div className="relative py-10 px-4 overflow-x-auto scrollbar-hide">
                    {/* Horizontal Signal Line (Desktop) */}
                    <div className="hidden xl:block absolute top-[68px] left-0 w-full h-[1px] bg-slate-200 dark:bg-border-dark z-0"></div>
                    {/* Vertical Signal Line (Mobile) */}
                    <div className="block xl:hidden absolute left-1/2 top-0 w-1 h-full bg-slate-200 dark:bg-border-dark -translate-x-1/2 z-0"></div>

                    {/* Blocks Container */}
                    <div className="relative z-10 flex flex-col xl:flex-row flex-wrap items-start justify-center gap-12 xl:gap-6 px-8 min-w-full xl:min-w-max">
                        {/* Signal Chain Blocks */}
                        {design.chain.map((comp, idx) => {
                            const type = comp.type?.toLowerCase() || 'default';
                            const cfg = typeConfig[type] || typeConfig.default;
                            const Icon = cfg.Icon;

                            // Snapshot logic for design view
                            let isEnabled = true;
                            if (activeSnapshot && activeSnapshot.active_blocks) {
                                isEnabled = activeSnapshot.active_blocks.some(name => {
                                    const n = name.toLowerCase();
                                    const bn = comp.name.toLowerCase();
                                    const bm = (comp.model || "").toLowerCase();
                                    // Stricter matching for AI-generated names
                                    return n === bn || n === bm || bn.includes(n) || bm.includes(n);
                                });
                            }

                            return (
                                <div key={idx} className={`flex flex-col items-center gap-2 group w-20 cursor-help transition-all duration-300 ${!isEnabled ? 'opacity-30 grayscale' : 'opacity-100'}`} title={comp.description}>
                                    {/* Stompbox Body */}
                                    <div className={`relative w-16 h-14 ${cfg.color} ${cfg.hoverColor} rounded-md border-b-4 ${cfg.borderColor} shadow-lg flex items-center justify-center transition-all transform hover:-translate-y-1 z-10 p-3`}>
                                        <Icon className="w-full h-full text-white drop-shadow-md select-none" />
                                    </div>

                                    {/* Labels */}
                                    <span className={`text-[9px] font-bold ${cfg.text} bg-white/90 dark:bg-background-dark/80 px-1.5 py-0.5 rounded border border-slate-300 dark:border-border-dark select-none uppercase tracking-tighter`}>
                                        {cfg.label}
                                    </span>
                                    <span className="text-[10px] text-slate-600 dark:text-gray-300 truncate w-full text-center font-medium px-1">
                                        {comp.name}
                                    </span>
                                </div>
                            );
                        })}
                    </div>
                </div>
            </div>

            {onGenerate && (
                <button
                    onClick={onGenerate}
                    className="self-end px-4 py-2 bg-primary/10 hover:bg-primary/20 text-primary text-xs font-bold rounded-lg border border-primary/20 hover:border-primary/40 transition-all flex items-center gap-2 group shadow-sm"
                >
                    <span className="material-symbols-outlined text-sm group-hover:animate-pulse">auto_fix_high</span>
                    Build this Rig
                </button>
            )}
        </div>
    );
};

export default DesignVisualizer;
