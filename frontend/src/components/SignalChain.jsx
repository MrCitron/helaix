import React, { useState } from 'react';
import { getIconForBlock, getBlockColor } from './IconLibrary';

const BlockParameters = ({ block, blockKey, color, onClose }) => {
    // Filter out metadata and system properties
    const getVisibleParams = (block) => {
        return Object.entries(block).filter(([key]) => {
            return !key.startsWith('@') &&
                key !== 'model' &&
                key !== 'type' &&
                typeof block[key] !== 'object';
        });
    };

    const params = getVisibleParams(block);

    return (
        <div className="p-6">
            <div className="flex items-start justify-between mb-4">
                <div>
                    <h4 className="text-slate-900 dark:text-white font-bold text-sm flex items-center gap-2">
                        <div className="w-1.5 h-3 rounded-full" style={{ backgroundColor: color }}></div>
                        {block["@model"].replace('HD2_', '').replace('VIC_', '')}
                    </h4>
                    <p className="text-[10px] text-slate-500 dark:text-text-muted font-mono mt-0.5 tracking-wider">{blockKey.toUpperCase()}</p>
                </div>
                <div className="flex gap-2">
                    <span className="text-[9px] px-2 py-0.5 rounded bg-slate-100 dark:bg-surface-dark text-primary border border-primary/20 font-bold uppercase">{block['@stereo'] ? 'Stereo' : 'Mono'}</span>
                    <button
                        onClick={onClose}
                        className="text-slate-500 dark:text-text-muted hover:text-slate-900 dark:hover:text-white transition-colors"
                    >
                        <span className="material-symbols-outlined text-sm">close</span>
                    </button>
                </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {params.length > 0 ? (
                    params.map(([pKey, pVal]) => (
                        <div key={pKey} className="bg-white dark:bg-background-dark border border-slate-300 dark:border-border-dark rounded-lg p-2.5 flex flex-col gap-1 hover:border-slate-400 dark:hover:border-[#3b4f54] transition-colors">
                            <span className="text-[10px] text-slate-500 dark:text-text-muted font-medium uppercase tracking-tight">{pKey}</span>
                            <div className="flex items-center justify-between">
                                <span className="text-slate-900 dark:text-white font-mono text-xs">
                                    {typeof pVal === 'number' ? pVal.toFixed(2) : String(pVal)}
                                </span>
                                {typeof pVal === 'number' && (
                                    <div className="w-12 h-1 bg-slate-300 dark:bg-border-dark rounded-full overflow-hidden">
                                        <div
                                            className="h-full bg-primary"
                                            style={{ width: `${Math.min(100, Math.max(0, pVal * 100))}%`, backgroundColor: color }}
                                        ></div>
                                    </div>
                                )}
                            </div>
                        </div>
                    ))
                ) : (
                    <div className="col-span-full py-4 text-center text-xs text-slate-500 dark:text-text-muted italic">
                        No adjustable parameters for this block.
                    </div>
                )}
            </div>
        </div>
    );
};

const SignalChain = ({ dspBlocks }) => {
    const [expandedBlock, setExpandedBlock] = useState(null);

    const toggleBlock = (key) => {
        setExpandedBlock(expandedBlock === key ? null : key);
    };

    return (
        <div className="w-full transition-all duration-300">
            {/* Scroll Container for the Chain */}
            <div className="w-full py-6 px-4 overflow-x-auto lg:overflow-visible scrollbar-hide">
                <div className="relative flex flex-col lg:flex-row items-center justify-center min-w-full lg:min-w-max gap-12 lg:gap-8 px-8 py-8 lg:py-0 h-auto lg:h-24">
                    {/* Horizontal Signal Line (Desktop) */}
                    <div className="hidden lg:block absolute top-1/2 left-0 w-full h-0.5 bg-slate-300 dark:bg-border-dark -translate-y-1/2 z-0"></div>
                    {/* Vertical Signal Line (Mobile) */}
                    <div className="block lg:hidden absolute left-1/2 top-0 w-0.5 h-full bg-slate-300 dark:bg-border-dark -translate-x-1/2 z-0"></div>

                    {dspBlocks.map(([key, block]) => {
                        const Icon = getIconForBlock(block);
                        const color = getBlockColor(block);
                        const isExpanded = expandedBlock === key;

                        return (
                            <React.Fragment key={key}>
                                <div className="relative z-10 flex flex-col items-center">
                                    {/* Connector Dot */}
                                    <div className="absolute lg:top-1/2 top-0 lg:-left-4 left-1/2 w-2 h-2 rounded-full bg-slate-400 dark:bg-[#3b4f54] -translate-y-4 lg:-translate-y-1/2 -translate-x-1/2 lg:translate-x-0"></div>

                                    {/* Block Node */}
                                    <button
                                        onClick={() => toggleBlock(key)}
                                        className={`group relative w-16 h-16 rounded-xl flex items-center justify-center transition-all duration-300 border-2
                                            ${isExpanded
                                                ? 'scale-110 shadow-[0_0_20px_rgba(19,200,236,0.2)] border-primary bg-white dark:bg-background-dark'
                                                : 'hover:scale-105 border-slate-300 dark:border-border-dark hover:border-slate-400 dark:hover:border-[#3b4f54] bg-slate-100 dark:bg-surface-dark'
                                            }`}
                                        style={{
                                            borderColor: isExpanded ? color : undefined,
                                            backgroundColor: isExpanded ? `${color}15` : undefined
                                        }}
                                    >
                                        <Icon className={`w-8 h-8 transition-colors ${isExpanded ? 'text-primary' : 'text-slate-600 dark:text-text-secondary group-hover:text-slate-900 dark:group-hover:text-white'}`} style={{ color: isExpanded ? color : undefined }} />

                                        {/* Status LED */}
                                        <div className={`absolute top-1 right-1 w-1.5 h-1.5 rounded-full ${block['@enabled'] !== false ? 'bg-primary shadow-[0_0_5px_#13c8ec]' : 'bg-red-900'}`}></div>
                                    </button>

                                    {/* Label */}
                                    <div className="mt-2 flex flex-col items-center">
                                        <span className={`text-[10px] font-bold truncate max-w-[80px] transition-colors ${isExpanded ? 'text-slate-900 dark:text-white' : 'text-slate-600 dark:text-text-muted'}`}>
                                            {block["@model"].replace('HD2_', '').replace('VIC_', '')}
                                        </span>
                                    </div>
                                </div>

                                {/* Inline Expanded Area (Vertical Mode Only) */}
                                <div className={`lg:hidden w-full relative z-10 overflow-hidden transition-all duration-300 ease-in-out ${isExpanded ? 'max-h-[800px] opacity-100 mt-4 border border-slate-300 dark:border-border-dark bg-white dark:bg-[#0b1011] rounded-xl shadow-2xl' : 'max-h-0 opacity-0'}`}>
                                    {isExpanded && (
                                        <BlockParameters
                                            block={block}
                                            blockKey={key}
                                            color={color}
                                            onClose={() => setExpandedBlock(null)}
                                        />
                                    )}
                                </div>
                            </React.Fragment>
                        );
                    })}

                    {/* End Cap */}
                    <div className="relative z-10 w-2 h-2 rounded-full bg-[#3b4f54] lg:mt-0 -mt-4"></div>
                </div>
            </div>

            {/* Bottom Expanded Area (Horizontal Mode Only) */}
            <div className={`hidden lg:block relative z-10 overflow-hidden transition-all duration-300 ease-in-out ${expandedBlock ? 'max-h-[500px] opacity-100 border-t border-[#283639] bg-[#0b1011]' : 'max-h-0 opacity-0'}`}>
                {expandedBlock && (() => {
                    const block = dspBlocks.find(([k]) => k === expandedBlock)[1];
                    const color = getBlockColor(block);
                    return (
                        <BlockParameters
                            block={block}
                            blockKey={expandedBlock}
                            color={color}
                            onClose={() => setExpandedBlock(null)}
                        />
                    );
                })()}
            </div>
        </div>
    );
};

export default SignalChain;
