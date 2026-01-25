import React, { useState } from 'react';
import { getIconForBlock, getBlockColor } from './IconLibrary';

const BlockParameters = ({ block, blockKey, color, activeSnapshot, preset, onClose, rigTotal }) => {
    // Determine the current value accounting for snapshot overrides
    const getParamValue = (pKey, baseVal) => {
        if (!preset || !preset.data || !preset.data.tone || !activeSnapshot) return baseVal;
        const snapIndex = activeSnapshot.index ?? 0;
        const snap = preset.data.tone[`snapshot${snapIndex}`];
        if (!snap || !snap.controllers) return baseVal;

        // Controllers are in dsp0 or dsp1
        const dspKey = block?._dsp || 'dsp0';
        const internalId = block?._id || blockKey;
        const ctrls = snap.controllers[dspKey] || {};
        const blockCtrls = ctrls[internalId];
        if (blockCtrls && blockCtrls[pKey] !== undefined) {
            return blockCtrls[pKey]["@value"] ?? baseVal;
        }
        return baseVal;
    };

    // Check if a parameter is controlled by snapshot (Controller 9)
    const isSnapshotControlled = (pKey) => {
        if (!preset || !preset.data || !preset.data.tone || !preset.data.tone.controller) return false;
        // Search the specific DSP path for controller data
        const dspKey = block?._dsp || 'dsp0';
        const internalId = block?._id || blockKey;
        const ctrls = preset.data.tone.controller[dspKey] || {};
        const blockCtrls = ctrls[internalId];
        if (!blockCtrls) return false;
        const pCtrl = blockCtrls[pKey];
        return pCtrl && pCtrl["@controller"] === 9;
    };

    const getVisibleParams = (block) => {
        const visible = Object.entries(block).filter(([key]) => {
            return !key.startsWith('@') &&
                key !== 'model' &&
                key !== 'type' &&
                key !== 'hasParamShift' &&
                key !== 'snapshots' &&
                key.indexOf('_') !== 0 &&
                typeof block[key] !== 'object';
        });

        if (block["@type"] === 'variax') {
            visible.push(["Model", block["@model"]]);
        }

        if ((block["@model"] === 'Variax' || block["@type"] === 'variax') && block["@variax_customtuning"] === true) {
            const strings = [
                { key: "@variax_str1tuning", label: "Str 1 (High E)" },
                { key: "@variax_str2tuning", label: "Str 2 (B)" },
                { key: "@variax_str3tuning", label: "Str 3 (G)" },
                { key: "@variax_str4tuning", label: "Str 4 (D)" },
                { key: "@variax_str5tuning", label: "Str 5 (A)" },
                { key: "@variax_str6tuning", label: "Str 6 (Low E)" },
            ];
            strings.forEach(str => {
                if (block[str.key] !== undefined) {
                    visible.push([str.label, block[str.key]]);
                }
            });
        }
        return visible;
    };

    const params = getVisibleParams(block);

    // DSP Calculation
    const dspMap = preset?.data?.meta?.dsp_map || {};
    const blockModel = block["@model"];
    const dspCost = dspMap[blockModel] || 0;

    // Names: Clean format for internal model names
    const formatHelixName = (m, type) => {
        if (!m) return "Block";
        if (type === 'variax') return m;
        return m.replace('HD2_', '').replace('VIC_', '').replace('L6SPB_', '').replace('L6C_', '')
            .replace(/_/g, ' ').replace(/([A-Z])/g, ' $1').trim();
    };
    const helixName = formatHelixName(block["@model"] || block.model, block["@type"]);

    return (
        <div className="p-6">
            <div className="flex items-start justify-between mb-4">
                <div>
                    <div className="flex items-center gap-2 mb-1">
                        <div className="w-1.5 h-3 rounded-full" style={{ backgroundColor: color }}></div>
                        <h4 className="text-slate-900 dark:text-white font-bold text-sm">
                            {helixName}
                        </h4>
                    </div>
                    <div className="flex items-center gap-2">
                        <p className="text-[10px] text-slate-500 dark:text-text-muted font-mono tracking-wider">{block._dsp?.toUpperCase() || "DSP0"} - {block._id?.toUpperCase() || blockKey.toUpperCase()}</p>
                        {dspCost > 0 && (
                            <span className="text-[9px] px-1.5 py-0 bg-primary/10 text-primary border border-primary/20 rounded font-black">
                                {dspCost.toFixed(1)} / {rigTotal?.toFixed(1) || "0.0"}
                            </span>
                        )}
                    </div>
                </div>
                <div className="flex gap-2">
                    <span className="text-[9px] px-2 py-0.5 rounded bg-slate-100 dark:bg-surface-dark text-primary border border-primary/20 font-bold uppercase">{block['@stereo'] || block.stereo ? 'Stereo' : 'Mono'}</span>
                    <button onClick={onClose} className="text-slate-500 dark:text-text-muted hover:text-slate-900 dark:hover:text-white transition-colors">
                        <span className="material-symbols-outlined text-sm">close</span>
                    </button>
                </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {params.length > 0 ? (
                    params.map(([pKey, pVal]) => {
                        const currentVal = getParamValue(pKey, pVal);
                        const isControlled = isSnapshotControlled(pKey);
                        return (
                            <div key={pKey} className={`bg-white dark:bg-background-dark border rounded-lg p-2.5 flex flex-col gap-1 transition-all ${isControlled ? 'border-blue-500/50 shadow-[0_0_10px_rgba(59,130,246,0.1)]' : 'border-slate-300 dark:border-border-dark hover:border-slate-400 dark:hover:border-[#3b4f54]'}`}>
                                <div className="flex justify-between items-center">
                                    <span className="text-[10px] text-slate-500 dark:text-text-muted font-medium uppercase tracking-tight">{pKey}</span>
                                    {isControlled && (
                                        <div className="flex items-center gap-1 bg-blue-500/10 px-1 rounded">
                                            <div className="w-1 h-1 rounded-full bg-blue-500 animate-pulse"></div>
                                            <span className="text-blue-500 text-[8px] font-black tracking-tighter uppercase whitespace-nowrap">SNAP</span>
                                        </div>
                                    )}
                                </div>
                                <div className="flex items-center justify-between">
                                    <span className={`font-mono text-xs ${isControlled ? 'text-blue-500 font-black' : 'text-slate-900 dark:text-white'}`}>
                                        {typeof currentVal === 'number' ? currentVal.toFixed(2) : String(currentVal)}
                                    </span>
                                    {typeof currentVal === 'number' && (
                                        <div className="w-12 h-1 bg-slate-300 dark:bg-border-dark rounded-full overflow-hidden">
                                            <div className="h-full bg-primary" style={{ width: `${Math.min(100, Math.max(0, currentVal * 100))}%`, backgroundColor: isControlled ? '#3b82f6' : color }}></div>
                                        </div>
                                    )}
                                </div>
                            </div>
                        );
                    })
                ) : (
                    <div className="col-span-full py-4 text-center text-xs text-slate-500 dark:text-text-muted italic">No adjustable parameters.</div>
                )}
            </div>
        </div>
    );
};

const SignalChain = ({ dspBlocks, activeSnapshot, preset }) => {
    const [expandedBlock, setExpandedBlock] = useState(null);

    const toggleBlock = (key) => {
        setExpandedBlock(expandedBlock === key ? null : key);
    };

    // Total DSP Sum
    const dspMap = preset?.data?.meta?.dsp_map || {};
    const totalDSP = dspBlocks.reduce((sum, [_, b]) => sum + (dspMap[b["@model"]] || 0), 0);

    return (
        <div className="w-full transition-all duration-300">
            <div className="flex justify-start px-8 pt-4">
                <div className="flex items-center gap-2 bg-slate-800/40 px-2 py-0.5 rounded-full border border-slate-700/50">
                    <span className="material-symbols-outlined text-[10px] text-primary">analytics</span>
                    <span className="text-[9px] font-bold text-slate-400 uppercase tracking-widest">Estimated DSP usage: </span>
                    <span className={`text-[10px] font-black ${totalDSP > 150 ? 'text-red-400' : totalDSP > 90 ? 'text-orange-400' : 'text-primary'}`}>
                        {totalDSP.toFixed(1)}
                    </span>
                </div>
            </div>

            <div className="w-full py-10 px-4 overflow-x-auto xl:overflow-visible scrollbar-hide relative">
                {/* Unified Signal Line (Desktop) */}
                <div className="hidden xl:block absolute top-[68px] left-0 w-full h-[1px] bg-slate-300 dark:bg-border-dark z-0"></div>
                {/* Unified Signal Line (Mobile) */}
                <div className="block xl:hidden absolute left-1/2 top-0 w-1 h-full bg-slate-300 dark:bg-border-dark -translate-x-1/2 z-0"></div>

                <div className="relative z-10 flex flex-col xl:flex-row items-start justify-center min-w-full xl:min-w-max gap-12 xl:gap-6 px-8 py-4 xl:py-0 h-auto xl:h-28">
                    {dspBlocks.map(([key, block]) => {
                        const Icon = getIconForBlock(block);
                        const color = getBlockColor(block);
                        const isExpanded = expandedBlock === key;

                        // Snapshot Logic (Dimming & Shifts)
                        let isEnabledInSnapshot = block["@enabled"] !== false; // Baseline: block's own state

                        // 1. Technical Bypass Lookup (Helix Chain) - ABSOLUTE TRUTH
                        if (preset && preset.data && preset.data.tone && activeSnapshot) {
                            const snapIndex = activeSnapshot.index ?? 0;
                            const snap = preset.data.tone[`snapshot${snapIndex}`];

                            if (snap && snap.blocks) {
                                // Use the source DSP and original ID for lookup
                                const dspKey = block._dsp || 'dsp0';
                                const internalId = block._id || key;
                                const dspMap = snap.blocks[dspKey] || {};
                                if (dspMap[internalId] !== undefined) {
                                    isEnabledInSnapshot = dspMap[internalId] === true;
                                }
                            }
                        }
                        // 2. Abstract String Lookup (Design Chain fallback)
                        else if (activeSnapshot && activeSnapshot.active_blocks) {
                            const bName = (block.name || block["@name"] || "").toLowerCase();
                            const bModel = (block.model || block["@model"] || "").toLowerCase();
                            isEnabledInSnapshot = activeSnapshot.active_blocks.some(name => {
                                const n = name.toLowerCase();
                                return n === bName || n === bModel || bName.includes(n) || bModel.includes(n);
                            });
                        }

                        // Labels: Clean format for internal model names
                        const formatHelixName = (m, type) => {
                            if (!m) return "Block";
                            if (type === 'variax') return m;
                            return m.replace('HD2_', '').replace('VIC_', '').replace('L6SPB_', '').replace('L6C_', '')
                                .replace(/_/g, ' ').replace(/([A-Z])/g, ' $1').trim();
                        };
                        const helixName = formatHelixName(block["@model"] || block.model, block["@type"]);

                        // Infer STOMP label: Fix for Cabinets (Type 2) vs Amps (Type 1)
                        const modelStr = (block["@model"] || "").toLowerCase();
                        const bType = block["@type"];
                        let label = 'FX';

                        const isCab = bType === 4 || modelStr.includes("cab") || modelStr.includes("micir");
                        const isAmp = !isCab && (bType === 1 || bType === 2 || bType === 3);

                        if (isAmp) label = 'AMP';
                        else if (isCab) label = 'CAB';
                        else if (modelStr.includes("dist") || modelStr.includes("kinky") || modelStr.includes("scream") || modelStr.includes("deez")) label = 'DIST';
                        else if (modelStr.includes("delay")) label = 'DLY';
                        else if (modelStr.includes("reverb")) label = 'REV';
                        else if (modelStr.includes("comp") || modelStr.includes("dynamics")) label = 'DYN';
                        else if (modelStr.includes("mod") || modelStr.includes("chorus") || modelStr.includes("tremolo") || modelStr.includes("poly")) label = 'MOD';
                        else if (bType === "variax") label = 'VARX';

                        return (
                            <React.Fragment key={key}>
                                <div className={`flex flex-col items-center gap-2 group transition-all duration-300 ${!isEnabledInSnapshot ? 'opacity-30 grayscale' : 'opacity-100'}`}>
                                    <button
                                        onClick={() => toggleBlock(key)}
                                        className={`relative w-16 h-14 rounded-md border-b-4 shadow-lg flex items-center justify-center transition-all transform hover:-translate-y-1 z-10 p-3
                                            ${isExpanded ? 'scale-110 -translate-y-1 shadow-xl' : ''}`}
                                        style={{
                                            backgroundColor: color,
                                            borderColor: `${color}cc`,
                                            boxShadow: isExpanded ? `0 0 20px ${color}40` : undefined
                                        }}
                                    >
                                        <Icon className="w-full h-full text-white drop-shadow-md select-none" />

                                        {/* Snapshot Mod Markers (kept subtle) */}
                                        {activeSnapshot && block.hasParamShift && (
                                            <div className="absolute -bottom-1 -right-1 w-2.5 h-2.5 bg-blue-500 rounded-full border-2 border-white dark:border-background-dark shadow-[0_0_5px_rgba(59,130,246,0.5)] z-20"></div>
                                        )}
                                    </button>

                                    {/* Labels aligned with Real Chain */}
                                    <span className="text-[9px] font-bold text-slate-500 dark:text-text-muted bg-white/90 dark:bg-background-dark/80 px-1.5 py-0.5 rounded border border-slate-300 dark:border-border-dark select-none uppercase tracking-tighter">
                                        {label}
                                    </span>
                                    <span className={`text-[10px] truncate w-24 text-center font-medium px-1 transition-colors ${isExpanded ? 'text-primary' : 'text-slate-600 dark:text-gray-300'}`}>
                                        {helixName}
                                    </span>

                                    {/* Inline Expanded Area (Vertical/Mobile) */}
                                    <div className={`xl:hidden w-80 relative z-10 overflow-hidden transition-all duration-300 ease-in-out ${isExpanded ? 'max-h-[800px] opacity-100 mt-4 border border-slate-300 dark:border-border-dark bg-white dark:bg-[#0b1011] rounded-xl shadow-2xl' : 'max-h-0 opacity-0'}`}>
                                        {isExpanded && (
                                            <BlockParameters block={block} blockKey={key} color={color} activeSnapshot={activeSnapshot} preset={preset} onClose={() => setExpandedBlock(null)} rigTotal={totalDSP} />
                                        )}
                                    </div>
                                </div>
                            </React.Fragment>
                        );
                    })}
                </div>
            </div>

            {!expandedBlock && (
                <div className="flex justify-center w-full mt-2">
                    <p className="text-[9px] text-[#586e75] italic pb-2 uppercase tracking-widest opacity-60 animate-in fade-in duration-500">Click a block to view parameters</p>
                </div>
            )}

            {/* Bottom Expanded Area (Horizontal/Desktop) */}
            <div className={`hidden xl:block relative z-10 overflow-hidden transition-all duration-300 ease-in-out ${expandedBlock ? 'max-h-[500px] opacity-100 border-t border-[#283639] bg-[#0b1011]' : 'max-h-0 opacity-0'}`}>
                {expandedBlock && (() => {
                    const blk = dspBlocks.find(([k]) => k === expandedBlock)[1];
                    const clr = getBlockColor(blk);
                    return <BlockParameters block={blk} blockKey={expandedBlock} color={clr} activeSnapshot={activeSnapshot} preset={preset} onClose={() => setExpandedBlock(null)} rigTotal={totalDSP} />;
                })()}
            </div>
        </div>
    );
};

export default SignalChain;
