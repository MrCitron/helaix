import React, { useState } from 'react';
import { useI18n } from '../i18n';
import SignalChain from './SignalChain';

function PresetVisualizer({ preset, design, compact = false, activeSnapIdx: propsActiveSnapIdx, hideSelector = false }) {
    const { t } = useI18n();
    const [localActiveSnapIdx, setLocalActiveSnapIdx] = useState(0);

    const activeSnapIdx = propsActiveSnapIdx !== undefined ? propsActiveSnapIdx : localActiveSnapIdx;
    const setActiveSnapIdx = propsActiveSnapIdx !== undefined ? () => { } : setLocalActiveSnapIdx;

    if (!preset) return null;

    const snapshots = (design?.snapshots || []).map((s, i) => ({ ...s, index: i }));
    const activeSnapshot = snapshots.length > 0 ? { ...(snapshots[activeSnapIdx] || snapshots[0]), index: activeSnapIdx } : null;

    // 1. Unified Block Collection (DSP 0 + DSP 1)
    const getBlocksFromDSP = (dspKey) => {
        const dsp = preset.data.tone[dspKey] || {};
        return Object.entries(dsp)
            .filter(([key]) => key.startsWith('block'))
            .map(([key, block]) => {
                // Find which snapshots this block is active in
                const snapshotsActiveIn = snapshots
                    .filter(s => s.active_blocks?.some(name => {
                        const n = name.toLowerCase();
                        const bn = (block["@name"] || "").toLowerCase();
                        const model = (block["@model"] || "").toLowerCase();
                        return n === bn || n === model || n.includes(bn) || bn.includes(n) || n.includes(model);
                    }))
                    .map(s => s.name);

                // Technical Check: Does this block have snapshot-controlled parameters?
                const controllers = preset.data.tone.controller?.[dspKey]?.[key] || {};
                const hasTechnicalShift = Object.values(controllers).some(c => c["@controller"] === 9);

                // Abstract Check: Did the AI mention a parameter shift here?
                let hasDesignShift = false;
                if (activeSnapshot && activeSnapshot.params) {
                    const blockKeyName = block["@name"]?.toLowerCase() || "";
                    const modelName = block["@model"]?.toLowerCase() || "";
                    hasDesignShift = Object.keys(activeSnapshot.params).some(k => {
                        const lk = k.toLowerCase();
                        return lk === blockKeyName || lk === modelName || lk.includes(blockKeyName) || blockKeyName.includes(lk);
                    });
                }

                const hasParamShift = hasTechnicalShift || hasDesignShift;

                return [`${dspKey}_${key}`, { ...block, snapshots: snapshotsActiveIn, hasParamShift, _dsp: dspKey, _id: key }];
            });
    };

    const dspBlocks = [...getBlocksFromDSP('dsp0'), ...getBlocksFromDSP('dsp1')]
        .sort((a, b) => {
            // Sort Strategy: DSP0 is ALWAYS before DSP1 in a serial chain
            if (a[1]._dsp !== b[1]._dsp) {
                return a[1]._dsp === 'dsp0' ? -1 : 1;
            }
            // Same DSP: Sort by position
            const posA = a[1]['@position'] ?? 0;
            const posB = b[1]['@position'] ?? 0;
            return posA - posB || a[0].localeCompare(b[0]);
        });

    // 2. Resolve Variax for Active Snapshot (Standard Hardware + Virtual Fallback)
    let activeVariax = preset.data.tone.variax;
    if (activeSnapshot) {
        const snap = preset.data.tone[`snapshot${activeSnapIdx}`];
        const snapVariaxCtrl = snap?.controllers?.variax;

        if (snapVariaxCtrl) {
            // Technical Lookup: Map controller '@value' back to the visual properties
            const overrides = {};
            Object.keys(snapVariaxCtrl).forEach(k => {
                if (snapVariaxCtrl[k] && snapVariaxCtrl[k]["@value"] !== undefined) {
                    overrides[k] = snapVariaxCtrl[k]["@value"];
                }
            });
            activeVariax = { ...activeVariax, ...overrides };
        } else {
            // Legacy/Virtual Fallback
            const snapVariax = snap?.variax;
            if (snapVariax && Object.keys(snapVariax).length > 0) {
                activeVariax = { ...activeVariax, ...snapVariax };
            }
        }
    }

    // 3. Inject Virtual Variax Block if data or controllers exist
    const hasVariaxData = preset.data.tone.variax && preset.data.tone.variax["@variax_model"] !== 0;
    const hasVariaxControllers = preset.data.tone.controller?.variax && Object.keys(preset.data.tone.controller.variax).length > 0;

    if (hasVariaxData || hasVariaxControllers) {
        const modelId = activeVariax?.["@variax_model"] || 0;
        const variaxType = preset.data.meta?.variax_type || "jtv";
        const getName = (id) => {
            if (id === 0) return "Variax";
            const jtvBanks = ["Custom 1", "T-Model", "Spank", "Lester", "Special", "R-Billy", "Chime", "Semi", "Jazzbox", "Acoustic", "Reso", "Custom 2"];
            const shurikenBanks = ["Shuriken", "T-Model", "Spank", "Lester", "Acoustic", "Jazz", "World", "Twang", "User I", "User II", "User III", "User IV"];

            const banks = variaxType === "shuriken" ? shurikenBanks : jtvBanks;
            const bankIdx = Math.floor((id - 1) / 5);
            // JTV/Classic logic: Variant 5 is first ID in bank, Variant 1 is last ID
            const variant = 5 - ((id - 1) % 5);
            const bankName = banks[bankIdx] || "Variax";
            return `${bankName} ${variant}`;
        };

        const variaxBlock = {
            ...activeVariax,
            "@model": getName(modelId),
            "@type": "variax",
            "@name": "Variax",
            "@enabled": true
        };
        // Unshift to the very beginning (Input Stage)
        dspBlocks.unshift(["variax", variaxBlock]);
    }

    return (
        <div className={`space-y-4 w-full animate-in fade-in duration-300 ${compact ? '' : 'p-4'}`}>
            {!compact && (
                <div className="flex items-center gap-2 text-primary border-b border-[#283639] pb-3">
                    <span className="material-symbols-outlined">description</span>
                    <h2 className="text-lg font-bold text-white truncate">{preset.data.meta.name}</h2>
                </div>
            )}

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
                                    : 'bg-[#1a2b2e] text-slate-400 border-[#283639] hover:border-primary/40'
                                }`}
                        >
                            {snap.name.toUpperCase()}
                        </button>
                    ))}
                </div>
            )}

            {/* Signal Chain Visualization */}
            <div className="bg-[#0b1011] rounded-2xl border border-[#283639] py-4 shadow-inner relative overflow-hidden flex flex-col">
                <div className="flex justify-between items-center px-6 mb-2">
                    <div className="flex flex-col gap-0.5">
                        <h3 className="text-text-muted uppercase text-[10px] font-bold tracking-[0.2em] opacity-50">{t('chat.helixChain')}</h3>
                        {activeSnapshot && (
                            <span className="text-primary text-[9px] font-black uppercase tracking-tighter">
                                Snapshot: {activeSnapshot.name}
                            </span>
                        )}
                    </div>
                </div>

                <div className="w-full">
                    <SignalChain dspBlocks={dspBlocks} activeSnapshot={activeSnapshot} preset={preset} />
                </div>
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
