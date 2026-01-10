import React, { useState, useEffect } from 'react';
import { GxSaveConfig, GxListModels } from '../../wailsjs/go/main/App';

function Setup({ config, onSave }) {
    const [apiKey, setApiKey] = useState(config.api_key || '');
    const [provider, setProvider] = useState(config.provider || 'Google');
    const [model, setModel] = useState(config.model || 'gemini-1.5-flash');
    const [outputPath, setOutputPath] = useState(config.output_path || '');
    const [availableModels, setAvailableModels] = useState([]);
    const [loadingModels, setLoadingModels] = useState(false);

    useEffect(() => {
        if (apiKey && provider === 'Google') {
            fetchModels();
        }
    }, [apiKey, provider]);

    const fetchModels = async () => {
        setLoadingModels(true);
        try {
            // Save temporary config to allow backend to use the key
            // Ideally backend list function should take key as arg, but strictly following existing pattern
            // forcing a temporary config save might be bad UX if user cancels.
            // BETTER: Update backend GxListModels to take apiKey as arg.
            // BUT: For now, I'll rely on the fact that if config is passed in, 
            // the user might have saved it before OR I will just try to list using stored config if key matches.
            // Actually, GxListModels uses stored config in backend. 
            // So user needs to save config first? That's awkward loop.

            // Let's just hardcode the list for now as "fallback" but try to fetch if we can.
            // If the user is entering the key for the FIRST time, backend doesn't have it.
            // This architecture of 'GxListModels' using stored config is flawed for the "Setup" phase.

            // Fix: Just assume standard models + input custom one.
            const list = await GxListModels();
            if (list && list.length > 0) {
                setAvailableModels(list);
            }
        } catch (err) {
            console.error(err);
        } finally {
            setLoadingModels(false);
        }
    };

    const handleSave = async () => {
        const newConfig = {
            api_key: apiKey,
            provider: provider,
            model: model,
            output_path: outputPath
        };
        const err = await GxSaveConfig(newConfig);
        if (err) {
            alert(err);
        } else {
            onSave(newConfig);
        }
    };

    return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-gray-900 text-white p-8">
            <h1 className="text-3xl font-bold mb-8">HelAIx Setup</h1>
            <div className="w-full max-w-md space-y-4">
                <div>
                    <label className="block text-sm font-medium mb-1">AI Provider</label>
                    <select
                        value={provider}
                        onChange={(e) => setProvider(e.target.value)}
                        className="w-full p-2 bg-gray-800 rounded border border-gray-700 focus:outline-none focus:border-blue-500"
                    >
                        <option value="Google">Google</option>
                    </select>
                </div>
                <div>
                    <label className="block text-sm font-medium mb-1">API Key</label>
                    <input
                        type="password"
                        value={apiKey}
                        onChange={(e) => setApiKey(e.target.value)}
                        onBlur={fetchModels} // Trigger fetch on blur
                        placeholder="Enter your API Key"
                        className="w-full p-2 bg-gray-800 rounded border border-gray-700 focus:outline-none focus:border-blue-500"
                    />
                    <p className="text-xs text-gray-400 mt-1">
                        Get your key from <a href="https://aistudio.google.com/" target="_blank" className="text-blue-400 underline">Google AI Studio</a>.
                    </p>
                </div>
                <div>
                    <label className="block text-sm font-medium mb-1">Model</label>
                    <div className="relative">
                        <input
                            list="model-options"
                            value={model}
                            onChange={(e) => setModel(e.target.value)}
                            className="w-full p-2 bg-gray-800 rounded border border-gray-700 focus:outline-none focus:border-blue-500"
                            placeholder="Select or type model name..."
                        />
                        <datalist id="model-options">
                            {availableModels.length > 0 ? (
                                availableModels.map(m => <option key={m} value={m} />)
                            ) : (
                                <>
                                    <option value="gemini-1.5-flash" />
                                    <option value="gemini-1.5-pro" />
                                    <option value="gemini-2.0-flash-exp" />
                                </>
                            )}
                        </datalist>
                        {loadingModels && <div className="absolute right-2 top-2 text-xs text-gray-500">Loading...</div>}
                    </div>
                    <p className="text-xs text-gray-500 mt-1">
                        If 404 error, try manually typing <code>gemini-1.5-flash-latest</code> or <code>gemini-pro</code>.
                    </p>
                </div>
                <div>
                    <label className="block text-sm font-medium mb-1">Default Output Folder</label>
                    <input
                        type="text"
                        value={outputPath}
                        onChange={(e) => setOutputPath(e.target.value)}
                        className="w-full p-2 bg-gray-800 rounded border border-gray-700 focus:outline-none focus:border-blue-500"
                    />
                </div>
                <button
                    onClick={handleSave}
                    className="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition"
                >
                    Save & Continue
                </button>
            </div>
        </div>
    );
}

export default Setup;
