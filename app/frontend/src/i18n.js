import { useSyncExternalStore } from 'react';

const translations = {
    en: {
        appName: "HelAIx",
        tagline: "Preset Generator",
        loading: "Loading...",
        newChatName: "New Chat",
        nav: {
            newChat: "Craft A Preset",
            settings: "Settings",
            recent: "Recent",
            noHistory: "No history yet"
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
            activeDiscussion: "Active Discussion",
            refinedTechnicalPreset: "I've refined the technical preset based on your feedback.",
            retry: "Retry without duplicating your message",
            errors: {
                ia: "AI might make mistakes. Always check output levels before playing.",
                aiFailed: "AI failed: ",
                buildFailed: "Build failed: ",
                exportFailed: "Export failed: "
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
            providerGemini: "Google Gemini API",
            apiKey: "API Key",
            testConn: "Test Connection",
            keyHint: "Your key is stored locally and never shared.",
            apiKeyPlaceholder: "Enter your API key...",
            loadingModels: "Loading models...",
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
            language: "Language",
            deleteNoConfirm: "Delete without confirmation",
            deleteNoConfirmHint: "Skip the confirmation popup when deleting a chat.",
            defaultExpPedal: "Default Expression Pedal",
            defaultExpHint: "The assigned controller for Wah, Volume, and Pitch Wham by default.",
            expOptions: {
                none: "Nothing",
                exp1: "Exp 1",
                exp2: "Exp 2",
                exp3: "Exp 3"
            },
            variaxSection: "Variax Control",
            variaxEnabled: "Automatic Variax Control",
            variaxEnabledHint: "Automatically configure your Variax guitar's model and tuning.",
            variaxHardwareModel: "Variax Hardware Model",
            variaxHardwareHint: "Select your specific Variax guitar for accurate modeling.",
            providerGeminiHint: "Using ai.google.dev API with API key authentication",
            providerCodexHint: "Uses your local Codex CLI session and consumes the limits included with your ChatGPT Codex plan.",
            codexModel: "Codex model",
            codexDefaultModel: "Let Codex choose the default model",
            codexCatalogVisible: "Only models Codex marks visible are listed. gpt-reserve and codex-auto-review are excluded because Codex marks them hidden.",
            codexCatalogUnavailable: "The local Codex model catalog could not be loaded. Codex will choose the default model.",
            codexModelUnavailable: "Codex will choose the default model (this CLI does not report --model support)",
            codexSubscription: "Codex - ChatGPT subscription",
            codexExecutablePath: "Codex executable path",
            codexExecutablePlaceholder: "Auto-detect from PATH, Homebrew, npm, or Volta",
            codexLoginHint: "Run codex login in Terminal and complete the official ChatGPT browser sign-in. Do not use an API key for this provider.",
            codexTesting: "Testing Codex...",
            codexTestConnection: "Test connection (uses Codex plan quota)",
            codexCheckingStatus: "Checking Codex CLI status...",
            codexNotFound: "Codex CLI was not found. Install it or set its executable path below.",
            codexNotSignedIn: "Codex CLI is not signed in with ChatGPT. Run codex login in Terminal.",
            codexApiKeyAuth: "Codex CLI is using an API key. Sign out and use ChatGPT sign-in for this provider.",
            codexConnected: "Codex CLI is signed in with ChatGPT. Requests use your Codex plan limits.",
            codexDetected: "Detected:",
            multiDsp: "Multi-DSP (Path 1 & 2)",
            singleDsp: "Single-DSP (Path 1 Only)",
            languageEnglish: "English",
            languageFrench: "French",
            languageSpanish: "Spanish",
            variaxOptions: {
                none: "Don't Force",
                tModel: "T-Model",
                spank: "Spank",
                lester: "Lester",
                special: "Special",
                rbilly: "R-Billy",
                chime: "Chime",
                semi: "Semi",
                jazzbox: "Jazzbox",
                acoustic: "Acoustic",
                reso: "Reso/Other"
            }
        },
        visualizer: {
            buildRig: "Build this Rig",
            snapshotMode: "Mode",
            standardTuning: "Standard",
            snapshot: "Snapshot",
            inspectJson: "Inspect JSON",
            stereo: "Stereo",
            mono: "Mono",
            snapshotControl: "Snap",
            noAdjustableParameters: "No adjustable parameters.",
            estimatedDspUsage: "Estimated DSP usage:",
            clickBlockParameters: "Click a block to view parameters",
            block: "Block",
            model: "Model",
            strings: {
                highE: "Str 1 (High E)", b: "Str 2 (B)", g: "Str 3 (G)",
                d: "Str 4 (D)", a: "Str 5 (A)", lowE: "Str 6 (Low E)"
            }
        }
    },
    fr: {
        appName: "HelAIx",
        tagline: "Générateur de Presets",
        loading: "Chargement...",
        newChatName: "Nouveau Chat",
        nav: {
            newChat: "Créer un Preset",
            settings: "Paramètres",
            recent: "Récents",
            noHistory: "Aucun historique"
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
            activeDiscussion: "Discussion active",
            refinedTechnicalPreset: "J'ai affiné le preset technique selon vos retours.",
            retry: "Réessayer sans dupliquer votre message",
            errors: {
                ia: "L'IA peut faire des erreurs. Vérifiez toujours vos niveaux de sortie avant de jouer.",
                aiFailed: "Échec de l'IA : ",
                buildFailed: "Échec de la création : ",
                exportFailed: "Échec de l'exportation : "
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
            providerGemini: "API Google Gemini",
            apiKey: "Clé API",
            testConn: "Tester la connexion",
            keyHint: "Votre clé est stockée localement et n'est jamais partagée.",
            apiKeyPlaceholder: "Saisissez votre clé API...",
            loadingModels: "Chargement des modèles...",
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
            language: "Langue",
            deleteNoConfirm: "Supprimer sans confirmation",
            deleteNoConfirmHint: "Passer la fenêtre de confirmation lors de la suppression d'un chat.",
            defaultExpPedal: "Pédale d'Expression par Défaut",
            defaultExpHint: "Le contrôleur assigné par défaut pour les blocs Wah, Volume et Pitch Wham.",
            expOptions: {
                none: "Rien",
                exp1: "Exp 1",
                exp2: "Exp 2",
                exp3: "Exp 3"
            },
            variaxSection: "Contrôle Variax",
            variaxEnabled: "Contrôle Variax Automatique",
            variaxEnabledHint: "Configure automatiquement le modèle et l'accordage de votre guitare Variax.",
            variaxHardwareModel: "Modèle de Guitare Variax",
            variaxHardwareHint: "Sélectionnez votre guitare Variax spécifique pour un modelage précis.",
            providerGeminiHint: "Utilise l'API ai.google.dev avec authentification par clé API",
            providerCodexHint: "Utilise votre session locale Codex CLI et les limites incluses dans votre forfait ChatGPT Codex.",
            codexModel: "Modèle Codex",
            codexDefaultModel: "Laisser Codex choisir le modèle par défaut",
            codexCatalogVisible: "Seuls les modèles que Codex marque comme visibles sont listés. gpt-reserve et codex-auto-review sont exclus car Codex les marque comme masqués.",
            codexCatalogUnavailable: "Le catalogue local de modèles Codex n'a pas pu être chargé. Codex choisira le modèle par défaut.",
            codexModelUnavailable: "Codex choisira le modèle par défaut (cette CLI ne prend pas en charge --model)",
            codexSubscription: "Codex - abonnement ChatGPT",
            codexExecutablePath: "Chemin de l'exécutable Codex",
            codexExecutablePlaceholder: "Détection automatique depuis PATH, Homebrew, npm ou Volta",
            codexLoginHint: "Exécutez codex login dans le Terminal et terminez la connexion officielle ChatGPT dans le navigateur. N'utilisez pas de clé API pour ce fournisseur.",
            codexTesting: "Test de Codex...",
            codexTestConnection: "Tester la connexion (utilise le quota du forfait Codex)",
            codexCheckingStatus: "Vérification de l'état de Codex CLI...",
            codexNotFound: "Codex CLI est introuvable. Installez-le ou indiquez son chemin ci-dessous.",
            codexNotSignedIn: "Codex CLI n'est pas connecté avec ChatGPT. Exécutez codex login dans le Terminal.",
            codexApiKeyAuth: "Codex CLI utilise une clé API. Déconnectez-vous et utilisez la connexion ChatGPT pour ce fournisseur.",
            codexConnected: "Codex CLI est connecté avec ChatGPT. Les requêtes utilisent les limites de votre forfait Codex.",
            codexDetected: "Détecté :",
            multiDsp: "Multi-DSP (chemins 1 et 2)",
            singleDsp: "Mono-DSP (chemin 1 uniquement)",
            languageEnglish: "Anglais",
            languageFrench: "Français",
            languageSpanish: "Espagnol",
            variaxOptions: {
                none: "Ne pas forcer",
                tModel: "T-Model",
                spank: "Spank",
                lester: "Lester",
                special: "Special",
                rbilly: "R-Billy",
                chime: "Chime",
                semi: "Semi",
                jazzbox: "Jazzbox",
                acoustic: "Acoustic",
                reso: "Reso/Autre"
            }
        },
        visualizer: {
            buildRig: "Créer ce rig",
            snapshotMode: "Mode",
            standardTuning: "Standard",
            snapshot: "Snapshot",
            inspectJson: "Inspecter le JSON",
            stereo: "Stéréo",
            mono: "Mono",
            snapshotControl: "Snap",
            noAdjustableParameters: "Aucun paramètre réglable.",
            estimatedDspUsage: "Utilisation DSP estimée :",
            clickBlockParameters: "Cliquez sur un bloc pour voir les paramètres",
            block: "Bloc",
            model: "Modèle",
            strings: {
                highE: "Cord. 1 (mi aigu)", b: "Cord. 2 (si)", g: "Cord. 3 (sol)",
                d: "Cord. 4 (ré)", a: "Cord. 5 (la)", lowE: "Cord. 6 (mi grave)"
            }
        }
    },
    es: {
        appName: "HelAIx",
        tagline: "Generador de Presets",
        loading: "Cargando...",
        newChatName: "Nuevo chat",
        nav: {
            newChat: "Crear un Preset",
            settings: "Configuración",
            recent: "Recientes",
            noHistory: "Aún no hay historial"
        },
        chat: {
            placeholder: "Describe tu sonido o valida el preset...",
            aiName: "asistente helAIx",
            userName: "Tú",
            soundEngineerName: "asistente helAIx",
            presetEngineerName: "asistente helAIx",
            welcome: "¡Hola! Estoy listo para configurar tu pedalera. ¿Qué tipo de sonido buscas hoy?",
            welcomeHint: "Ejemplo: 'Un sonido limpio y cristalino con mucho reverb' o 'Un lead al estilo Gilmour de 1979'",
            proposedChain: "Esta es la cadena de señal propuesta:",
            helixChain: "Cadena Helix",
            realChain: "Cadena Real",
            readyStatus: "Listo",
            signalChainStatus: "Estado de la Cadena de Señal",
            generateBtn: "Crear este rig",
            exportBtn: "Exportar archivo .hlx",
            currentChat: "Chat actual",
            activeDiscussion: "Conversación activa",
            refinedTechnicalPreset: "He refinado el preset técnico según tus comentarios.",
            retry: "Reintentar sin duplicar tu mensaje",
            errors: {
                ia: "La IA puede cometer errores. Comprueba siempre los niveles de salida antes de tocar.",
                aiFailed: "Error de IA: ",
                buildFailed: "Error al crear: ",
                exportFailed: "Error al exportar: "
            },
            deleteChat: "Eliminar chat",
            deleteConfirmTitle: "¿Eliminar chat?",
            deleteConfirmMsg: "¿Seguro que deseas eliminar esta conversación? Esta acción no se puede deshacer.",
            deleteConfirmOk: "Eliminar",
            deleteConfirmCancel: "Cancelar"
        },
        exportModal: {
            title: "Preset generado correctamente",
            description: "El archivo de configuración se ha compilado y exportado. Ya puedes importarlo en tu procesador Helix.",
            fileSavedAt: "Archivo guardado en:",
            openFolder: "Abrir carpeta de destino",
            openFile: "Abrir carpeta",
            close: "Cerrar"
        },
        settings: {
            title: "Configuración",
            subtitle: "Configura tu proveedor de IA, administra las claves API locales y establece las preferencias de exportación.",
            aiSection: "Inteligencia Artificial",
            provider: "Proveedor de IA",
            model: "Modelo LLM",
            providerGemini: "API de Google Gemini",
            apiKey: "Clave API",
            testConn: "Probar conexión",
            keyHint: "Tu clave se guarda localmente y nunca se comparte.",
            apiKeyPlaceholder: "Introduce tu clave API...",
            loadingModels: "Cargando modelos...",
            exportSection: "Carpeta de exportación",
            browse: "Explorar",
            folderHint: "Los archivos generados se guardarán aquí automáticamente.",
            overwrite: "Sobrescribir archivos existentes",
            overwriteHint: "Reemplazar sin preguntar si ya existe un archivo.",
            openFolder: "Abrir carpeta tras exportar",
            openFolderHint: "Abrir automáticamente el explorador de archivos.",
            incrementalSave: "Sufijo numérico incremental",
            incrementalSaveHint: "Añade un número (p. ej., _1, _2) si el archivo ya existe en lugar de sobrescribirlo.",
            hardwareTarget: "Modelo Helix",
            hardwareHint: "Limita la complejidad del preset según la potencia de procesamiento de tu dispositivo.",
            statusOk: "Todos los sistemas funcionan correctamente",
            cancel: "Cancelar",
            save: "Guardar",
            interfaceSection: "Interfaz",
            language: "Idioma",
            deleteNoConfirm: "Eliminar sin confirmación",
            deleteNoConfirmHint: "Omitir la ventana de confirmación al eliminar un chat.",
            defaultExpPedal: "Pedal de expresión predeterminado",
            defaultExpHint: "El controlador asignado por defecto para Wah, volumen y Pitch Wham.",
            expOptions: {
                none: "Ninguno",
                exp1: "Exp 1",
                exp2: "Exp 2",
                exp3: "Exp 3"
            },
            variaxSection: "Control Variax",
            variaxEnabled: "Control Variax automático",
            variaxEnabledHint: "Configura automáticamente el modelo y la afinación de tu guitarra Variax.",
            variaxHardwareModel: "Modelo de guitarra Variax",
            variaxHardwareHint: "Selecciona tu guitarra Variax específica para obtener un modelado preciso.",
            providerGeminiHint: "Usa la API de ai.google.dev con autenticación mediante clave API",
            providerCodexHint: "Usa tu sesión local de Codex CLI y consume los límites incluidos en tu plan ChatGPT Codex.",
            codexModel: "Modelo Codex",
            codexDefaultModel: "Dejar que Codex elija el modelo predeterminado",
            codexCatalogVisible: "Solo se muestran los modelos que Codex marca como visibles. gpt-reserve y codex-auto-review se excluyen porque Codex los marca como ocultos.",
            codexCatalogUnavailable: "No se pudo cargar el catálogo local de modelos de Codex. Codex elegirá el modelo predeterminado.",
            codexModelUnavailable: "Codex elegirá el modelo predeterminado (esta CLI no informa compatibilidad con --model)",
            codexSubscription: "Codex - suscripción de ChatGPT",
            codexExecutablePath: "Ruta del ejecutable de Codex",
            codexExecutablePlaceholder: "Detección automática desde PATH, Homebrew, npm o Volta",
            codexLoginHint: "Ejecuta codex login en Terminal y completa el inicio de sesión oficial de ChatGPT en el navegador. No uses una clave API para este proveedor.",
            codexTesting: "Probando Codex...",
            codexTestConnection: "Probar conexión (usa cuota del plan Codex)",
            codexCheckingStatus: "Comprobando el estado de Codex CLI...",
            codexNotFound: "No se encontró Codex CLI. Instálalo o indica su ruta de ejecutable abajo.",
            codexNotSignedIn: "Codex CLI no ha iniciado sesión con ChatGPT. Ejecuta codex login en Terminal.",
            codexApiKeyAuth: "Codex CLI está usando una clave API. Cierra sesión y usa el inicio de sesión de ChatGPT para este proveedor.",
            codexConnected: "Codex CLI ha iniciado sesión con ChatGPT. Las solicitudes usan los límites de tu plan Codex.",
            codexDetected: "Detectado:",
            multiDsp: "Multi-DSP (rutas 1 y 2)",
            singleDsp: "DSP único (solo ruta 1)",
            languageEnglish: "Inglés",
            languageFrench: "Francés",
            languageSpanish: "Español",
            variaxOptions: {
                none: "No forzar",
                tModel: "T-Model",
                spank: "Spank",
                lester: "Lester",
                special: "Special",
                rbilly: "R-Billy",
                chime: "Chime",
                semi: "Semi",
                jazzbox: "Jazzbox",
                acoustic: "Acoustic",
                reso: "Reso/Otro"
            }
        },
        visualizer: {
            buildRig: "Crear este rig",
            snapshotMode: "Modo",
            standardTuning: "Estándar",
            snapshot: "Snapshot",
            inspectJson: "Inspeccionar JSON",
            stereo: "Estéreo",
            mono: "Mono",
            snapshotControl: "Snap",
            noAdjustableParameters: "No hay parámetros ajustables.",
            estimatedDspUsage: "Uso estimado de DSP:",
            clickBlockParameters: "Haz clic en un bloque para ver los parámetros",
            block: "Bloque",
            model: "Modelo",
            strings: {
                highE: "Cda. 1 (mi agudo)", b: "Cda. 2 (si)", g: "Cda. 3 (sol)",
                d: "Cda. 4 (re)", a: "Cda. 5 (la)", lowE: "Cda. 6 (mi grave)"
            }
        }
    }
};

const languageListeners = new Set();
const defaultChatNames = new Set(['New Chat', 'Nouveau Chat', 'Nuevo chat']);

const getLanguage = () => {
    const savedLanguage = localStorage.getItem('helAIx_lang');
    return translations[savedLanguage] ? savedLanguage : 'en';
};

const subscribeToLanguage = (listener) => {
    languageListeners.add(listener);
    return () => languageListeners.delete(listener);
};

export const isDefaultChatName = (name) => defaultChatNames.has(name);

export const useI18n = () => {
    const lang = useSyncExternalStore(subscribeToLanguage, getLanguage, () => 'en');

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
            localStorage.setItem('helAIx_lang', newLang);
            languageListeners.forEach((listener) => listener());
        }
    };

    return { t, lang, changeLang };
};
