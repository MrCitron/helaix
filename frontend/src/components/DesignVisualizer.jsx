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
    default: { Icon: HelixIcons.FX, color: 'bg-[#607D8B]', hoverColor: 'hover:bg-[#78909C]', borderColor: 'border-[#455A64]', text: 'text-[#B0BEC5]', label: 'FX' }
};

const DesignVisualizer = ({ design, onGenerate }) => {
    const { t } = useI18n();

    if (!design || !design.chain) return null;

    return (
        <div className="flex flex-col gap-6 w-full">
            <div className="relative w-full rounded-xl border border-slate-300 dark:border-[#3f5256] bg-white dark:bg-background-dark p-4 py-8 overflow-hidden">
                <h3 className="text-slate-500 dark:text-text-muted uppercase text-[10px] font-bold tracking-[0.2em] pl-2 mb-6 opacity-50">{t('chat.realChain')}</h3>
                {/* Horizontal Signal Line (Desktop) */}
                <div className="hidden lg:block absolute top-1/2 left-0 w-full h-1 bg-slate-300 dark:bg-border-dark -translate-y-1/2 z-0"></div>
                {/* Vertical Signal Line (Mobile) */}
                <div className="block lg:hidden absolute left-1/2 top-0 w-1 h-full bg-slate-300 dark:bg-border-dark -translate-x-1/2 z-0"></div>

                {/* Blocks Container */}
                <div className="relative z-10 flex flex-col lg:flex-row flex-wrap items-center justify-center gap-8 lg:gap-4">
                    {design.chain.map((comp, idx) => {
                        const type = comp.type?.toLowerCase() || 'default';
                        const cfg = typeConfig[type] || typeConfig.default;
                        const Icon = cfg.Icon;

                        return (
                            <div key={idx} className="flex flex-col items-center gap-2 group w-20 cursor-help" title={comp.description}>
                                {/* Stompbox Body */}
                                <div className={`relative w-16 h-14 ${cfg.color} ${cfg.hoverColor} rounded-md border-b-4 ${cfg.borderColor} shadow-lg flex items-center justify-center transition-all transform hover:-translate-y-1 z-10 p-3`}>
                                    <Icon className="w-full h-full text-white drop-shadow-md select-none" />
                                    {/* LED */}
                                    <div className="absolute top-1 right-1 w-1.5 h-1.5 bg-red-400 rounded-full shadow-[0_0_5px_rgba(239,68,68,0.8)]"></div>
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
