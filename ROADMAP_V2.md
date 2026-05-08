# SentiGraph 2.0 Roadmap - Advanced Market Intelligence

## 🚀 Phase 1: High-Frequency Depth Analysis (The "Whale" Scan)
*   **Harvester**: S'abonner au flux `btcusdt@depth20@100ms` de Binance.
*   **Analyst**: Calculer l'**Order Book Imbalance** (Volume total d'achat vs Volume total de vente sur les 20 premiers niveaux).
*   **UI**: Ajouter une jauge secondaire "Book Pressure" pour voir les murs d'ordres.

## 📰 Phase 2: Real-World News & Social Ingestion
*   **Nouveau Service**: `sentigraph-scout` (Python ou Node.js).
*   **Fonction**: Scraper l'API CryptoPanic et les recherches Twitter (X) pour "Bitcoin".
*   **Intégration**: Envoyer les titres réels à l'Analyst via Redis à la place des titres statiques actuels.

## 🧠 Phase 3: AI Contextual Awareness (Deep Reasoning)
*   **Injection de Contexte**: Envoyer au modèle Llama3 :
    *   Variation de prix (%) sur les 5 dernières minutes.
    *   Pression du carnet d'ordres (Phase 1).
    *   Le titre de la news.
*   **Résultat**: Une décision multi-facteurs (Achat/Vente/Attente) avec un score de confiance basé sur la confluence.

## 📊 Phase 4: Multi-Asset Dashboard & Corrélation
*   **Scalabilité**: Lancer des instances Harvester pour ETH et SOL.
*   **UI**: Sidebar pour switcher entre les actifs.
*   **Alerting**: Notifications push quand le sentiment IA diverge du prix (ex: News Bullish mais prix qui chute = Manipulation probable).

## 🐋 Phase 5: Liquidation Heatmap
*   **Analyse**: Suivre les liquidations massives en temps réel pour détecter les "short/long squeezes" imminents.
