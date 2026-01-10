import React from 'react';

// Mockup-inspired custom icons for HelAIx
export const HelixIcons = {
    // Amp (Briefcase/Head)
    Amp: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
            <rect x="5" y="8" width="14" height="10" rx="1.5" />
            <path d="M9 8V6a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2" />
        </svg>
    ),
    // Cab (Target with dot)
    Cab: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
            <circle cx="12" cy="12" r="8" />
            <circle cx="12" cy="12" r="2" fill="currentColor" />
        </svg>
    ),
    // Distortion (Zig-zag)
    Distortion: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" className={className}>
            <path d="M3 12h3l3-6 6 12 3-6h3" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
    ),
    // Dynamics (Spark/Lightning)
    Dynamics: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
            <path d="M13 3L6 13h5l-2 8 7-10h-5l2-8z" strokeLinecap="round" strokeLinejoin="round" fill="none" />
        </svg>
    ),
    // Delay (Target with ticks)
    Delay: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
            <circle cx="12" cy="12" r="7" />
            <circle cx="12" cy="12" r="3" />
            <path d="M12 2v3M12 19v3M2 12h3M19 12h3" opacity="0.6" />
        </svg>
    ),
    // Reverb (Cube)
    Reverb: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
            <path d="M12 3l8 4.5v9L12 21l-8-4.5v-9L12 3z" />
            <path d="M12 12l8-4.5M12 12v9M12 12L4 7.5" />
        </svg>
    ),
    // Modulation (Infinity/Wave)
    Modulation: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" className={className}>
            <path d="M7 12c0-3 2-5 5-5s5 2 5 5-2 5-5 5-5-2-5-5z" opacity="0.3" />
            <path d="M3 12c0-4 3-7 7-7s4 3 6 7 2 7 6 7 3-3 5-7" strokeLinecap="round" />
        </svg>
    ),
    // Wah (Pedal with bars)
    Wah: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
            <rect x="7" y="4" width="10" height="16" rx="2" />
            <path d="M9 8h6M9 12h6M9 16h6" opacity="0.8" />
        </svg>
    ),
    // Volume (Bars)
    Volume: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="0" className={className}>
            <rect x="4" y="14" width="3" height="4" fill="currentColor" rx="0.5" />
            <rect x="9" y="11" width="3" height="7" fill="currentColor" rx="0.5" />
            <rect x="14" y="8" width="3" height="10" fill="currentColor" rx="0.5" />
            <rect x="19" y="5" width="3" height="13" fill="currentColor" rx="0.5" />
        </svg>
    ),
    // EQ
    EQ: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
            <path d="M3 12h18" opacity="0.3" />
            <path d="M7 8v8M12 4v16M17 8v8" strokeLinecap="round" />
        </svg>
    ),
    // Default / FX
    FX: ({ className }) => (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
            <circle cx="12" cy="12" r="9" strokeDasharray="2 2" />
            <path d="M12 8v8M8 12h8" />
        </svg>
    )
};

export const getIconForBlock = (block) => {
    const model = block["@model"] || "";
    const type = block["@type"];

    if (type === 1 || type === 2 || type === 3) return HelixIcons.Amp;
    if (type === 4) return HelixIcons.Cab;

    if (model.includes("Dist") || model.includes("Kinky") || model.includes("Scream") || model.includes("Minotaur")) return HelixIcons.Distortion;
    if (model.includes("Delay")) return HelixIcons.Delay;
    if (model.includes("Reverb")) return HelixIcons.Reverb;
    if (model.includes("Comp") || model.includes("Dynamics")) return HelixIcons.Dynamics;
    if (model.includes("Mod") || model.includes("Tremolo") || model.includes("Chorus") || model.includes("Vibe") || model.includes("Rotary")) return HelixIcons.Modulation;
    if (model.includes("EQ")) return HelixIcons.EQ;
    if (model.includes("Wah")) return HelixIcons.Wah;
    if (model.includes("Volume") || model.includes("Pan")) return HelixIcons.Volume;

    return HelixIcons.FX;
};

export const getBlockColor = (block) => {
    const type = block["@type"];
    const model = block["@model"] || "";

    if (type === 1 || type === 2 || type === 3) return "#FF5252"; // Amp (Red)
    if (type === 4) return "#D32F2F"; // Cab (Dark Red)

    if (model.includes("Dist") || model.includes("Kinky") || model.includes("Scream")) return "#FF5722"; // Distortion (Orange)
    if (model.includes("Delay")) return "#4CAF50"; // Delay (Green)
    if (model.includes("Reverb")) return "#FF9800"; // Reverb (Orange/Yellow)
    if (model.includes("Comp") || model.includes("Dynamics")) return "#FFC107"; // Dynamics (Amber)
    if (model.includes("Mod") || model.includes("Tremolo") || model.includes("Chorus")) return "#2196F3"; // Modulation (Blue)
    if (model.includes("EQ")) return "#FFC107"; // EQ (Amber)
    if (model.includes("Wah")) return "#9C27B0"; // Wah (Purple)
    if (model.includes("Volume") || model.includes("Pan")) return "#00BCD4"; // Vol (Cyan)

    return "#607D8B"; // Default
};
