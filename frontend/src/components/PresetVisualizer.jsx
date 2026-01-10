import React from 'react';
import { useI18n } from '../i18n';
import SignalChain from './SignalChain';

function PresetVisualizer({ preset, compact = false }) {
    const { t } = useI18n();

    if (!preset) return null;

    const dspBlocks = Object.entries(preset.data.tone.dsp0)
        .filter(([key]) => key.startsWith('block'))
        .sort((a, b) => {
            // Sort by position if available, otherwise by key
            const posA = a[1]['@position'] ?? 0;
            const posB = b[1]['@position'] ?? 0;
            return posA - posB || a[0].localeCompare(b[0]);
        });

    return (
        <div className={`space-y-4 w-full animate-in fade-in duration-300 ${compact ? '' : 'p-4'}`}>
            {!compact && (
                <div className="flex items-center gap-2 text-primary border-b border-[#283639] pb-3">
                    <span className="material-symbols-outlined">description</span>
                    <h2 className="text-lg font-bold text-white truncate">{preset.data.meta.name}</h2>
                </div>
            )}

            {/* Signal Chain Visualization */}
            <div className="space-y-3 bg-[#0b1011] rounded-2xl border border-[#283639] py-4 shadow-inner">
                <h3 className="text-text-muted uppercase text-[10px] font-bold tracking-[0.2em] pl-6 mb-2 opacity-50">{t('chat.helixChain')}</h3>
                <SignalChain dspBlocks={dspBlocks} />
                <p className="text-center text-[9px] text-[#586e75] italic pb-2 pb-2">Click a block to view parameters</p>
            </div>

            {/* Raw JSON Debug */}
            <details className="group pt-1">
                <summary className="cursor-pointer text-text-muted hover:text-white text-[10px] uppercase font-bold tracking-widest list-none flex items-center justify-center gap-1 transition-colors opacity-40 hover:opacity-100">
                    <span className="material-symbols-outlined text-[14px] transition-transform group-open:rotate-90">chevron_right</span>
                    Inspect JSON
                </summary>
                <div className="mt-2 bg-[#0b1011] p-3 rounded-lg border border-[#283639] overflow-hidden shadow-inner">
                    <pre className="text-[10px] text-[#9db4b9] font-mono overflow-auto max-h-48 scrollbar-hide selection:bg-primary/20">
                        {JSON.stringify(preset, null, 2)}
                    </pre>
                </div>
            </details>
        </div>
    );
}

export default PresetVisualizer;
