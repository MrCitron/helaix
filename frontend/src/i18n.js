import { useState, useEffect } from 'react';

const translations = {
    en: {
        appName: "HelAIx",
        tagline: "Preset Generator",
        nav: {
            newChat: "Craft A Preset",
            settings: "Settings",
            recent: "Recent"
        },
        chat: {
            placeholder: "Describe your sound or validate preset...",
            aiName: "helAIx assistant",
            userName: "You",
            soundEngineerName: "helAIx assistant",
            presetEngineerName: "helAIx assistant",
            welcome: "Hello! I'm ready to configure your pedalboard. What kind of sound are you looking for today?",
            welcomeHint: "Example: 'A crystal clean sound with lots of reverb' or 'A Gilmour 1979 style lead'",
            proposedChain: "Here is the proposed signal chain:",
            helixChain: "Helix Chain",
            realChain: "Real Chain",
            readyStatus: "Ready",
            signalChainStatus: "Signal Chain Status",
            generateBtn: "Build this rig",
            exportBtn: "Export .hlx file",
            currentChat: "Current Chat",
            errors: {
                ia: "AI might make mistakes. Always check output levels before playing."
            },
            deleteChat: "Delete Chat",
            deleteConfirmTitle: "Delete Chat?",
            deleteConfirmMsg: "Are you sure you want to delete this conversation? This action cannot be undone.",
            deleteConfirmOk: "Delete",
            deleteConfirmCancel: "Cancel"
        },
        exportModal: {
            title: "Preset Generated Successfully",
            description: "The configuration file has been compiled and exported. You can now import it into your Helix processor.",
            fileSavedAt: "File saved at:",
            openFolder: "Open Target Folder",
            openFile: "Open Folder",
            close: "Close"
        },
        settings: {
            title: "Settings",
            subtitle: "Configure your AI provider, manage local API keys and set export preferences.",
            aiSection: "Artificial Intelligence",
            provider: "AI Provider",
            model: "LLM Model",
            apiKey: "API Key",
            testConn: "Test Connection",
            keyHint: "Your key is stored locally and never shared.",
            exportSection: "Export target folder",
            browse: "Browse",
            folderHint: "Generated files will be automatically saved here.",
            overwrite: "Overwrite existing files",
            overwriteHint: "Replace without asking if a file exists.",
            openFolder: "Open folder after export",
            openFolderHint: "Automatically open file explorer.",
            incrementalSave: "Incremental Number Suffix",
            incrementalSaveHint: "Add a number (e.g. _1, _2) if the file already exists instead of overwriting.",
            hardwareTarget: "Helix Model",
            hardwareHint: "Constrain preset complexity based on your device's processing power.",
            statusOk: "All systems operational",
            cancel: "Cancel",
            save: "Save",
            interfaceSection: "Interface",
            deleteNoConfirm: "Delete without confirmation",
            deleteNoConfirmHint: "Skip the confirmation popup when deleting a chat."
        }
    },
    fr: {
        appName: "HelAIx",
        tagline: "Générateur de Presets",
        nav: {
            newChat: "Créer un Preset",
            settings: "Paramètres",
            recent: "Récents"
        },
        chat: {
            placeholder: "Décrivez votre modification ou validez le preset...",
            aiName: "assistant helAIx",
            userName: "Vous",
            soundEngineerName: "assistant helAIx",
            presetEngineerName: "assistant helAIx",
            welcome: "Bonjour ! Je suis prêt à configurer votre pédalier. Quel type de son cherchez-vous à créer aujourd'hui ?",
            welcomeHint: "Exemple : 'Un son clean cristallin avec beaucoup de reverb' ou 'Un son lead style Gilmour 1979'",
            proposedChain: "Voici la chaîne de signal proposée :",
            helixChain: "Helix Chain",
            realChain: "Real Chain",
            readyStatus: "Prêt",
            signalChainStatus: "Signal Chain Status",
            generateBtn: "Générer ce rig",
            exportBtn: "Exporter le fichier .hlx",
            currentChat: "Chat en cours",
            errors: {
                ia: "L'IA peut faire des erreurs. Vérifiez toujours vos niveaux de sortie avant de jouer."
            },
            deleteChat: "Supprimer le Chat",
            deleteConfirmTitle: "Supprimer le Chat ?",
            deleteConfirmMsg: "Êtes-vous sûr de vouloir supprimer cette conversation ? Cette action est irréversible.",
            deleteConfirmOk: "Supprimer",
            deleteConfirmCancel: "Annuler"
        },
        exportModal: {
            title: "Preset généré avec succès",
            description: "Le fichier de configuration a été compilé et exporté. Vous pouvez maintenant l'importer dans votre pédalier Helix.",
            fileSavedAt: "Fichier enregistré sous :",
            openFolder: "Ouvrir le dossier cible",
            openFile: "Ouvrir le dossier",
            close: "Fermer"
        },
        settings: {
            title: "Paramètres",
            subtitle: "Gérez votre configuration AI et vos préférences d'exportation.",
            aiSection: "Intelligence Artificielle",
            provider: "Fournisseur d'IA",
            model: "Modèle LLM",
            apiKey: "Clé API",
            testConn: "Tester la connexion",
            keyHint: "Votre clé est stockée localement et n'est jamais partagée.",
            exportSection: "Dossier d'exportation cible",
            browse: "Parcourir",
            folderHint: "Les fichiers générés seront automatiquement sauvegardés ici.",
            overwrite: "Écraser les fichiers existants",
            overwriteHint: "Remplacer sans demander si un fichier existe.",
            openFolder: "Ouvrir le dossier après export",
            openFolderHint: "Ouvre automatiquement l'explorateur de fichiers.",
            incrementalSave: "Suffixe Numérique Incrémental",
            incrementalSaveHint: "Ajoute un numéro (ex: _1, _2) si le fichier existe déjà au lieu de l'écraser.",
            hardwareTarget: "Modèle Helix",
            hardwareHint: "Limite la complexité des presets selon la puissance de votre appareil.",
            statusOk: "Tous les systèmes sont opérationnels",
            cancel: "Annuler",
            save: "Enregistrer",
            interfaceSection: "Interface",
            deleteNoConfirm: "Supprimer sans confirmation",
            deleteNoConfirmHint: "Passer la fenêtre de confirmation lors de la suppression d'un chat."
        }
    }
};

export const useI18n = () => {
    // Basic implementation: check localStorage or default to 'en'
    const [lang, setLang] = useState(localStorage.getItem('helAIx_lang') || 'en');

    const t = (path) => {
        const keys = path.split('.');
        let result = translations[lang];
        for (const key of keys) {
            if (result[key]) result = result[key];
            else return path; // Fallback to path name
        }
        return result;
    };

    const changeLang = (newLang) => {
        if (translations[newLang]) {
            setLang(newLang);
            localStorage.setItem('helAIx_lang', newLang);
        }
    };

    return { t, lang, changeLang };
};
