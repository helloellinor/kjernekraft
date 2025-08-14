package handlers

import (
	"html/template"
	"kjernekraft/models"
	"net/http"
	"strconv"
)

// KlippekortPageHandler serves the klippekort two-step selection page
func KlippekortPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Klippekort - Kjernekraft</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f5f5f5;
            color: #333;
        }
        .header {
            background-color: #007cba;
            color: white;
            padding: 1rem 2rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header h1 {
            font-size: 1.5rem;
            font-weight: 600;
        }
        .nav {
            background-color: white;
            border-bottom: 1px solid #e0e0e0;
            padding: 0;
        }
        .nav-list {
            display: flex;
            list-style: none;
            max-width: 1200px;
            margin: 0 auto;
        }
        .nav-item {
            border-right: 1px solid #e0e0e0;
        }
        .nav-item:last-child {
            border-right: none;
        }
        .nav-link {
            display: block;
            padding: 1rem 2rem;
            text-decoration: none;
            color: #333;
            font-weight: 500;
            transition: background-color 0.2s;
        }
        .nav-link:hover, .nav-link.active {
            background-color: #f0f8ff;
            color: #007cba;
        }
        .main {
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
        }
        .page-title {
            font-size: 2rem;
            margin-bottom: 1rem;
            color: #333;
        }
        .page-description {
            font-size: 1.1rem;
            color: #666;
            margin-bottom: 3rem;
            text-align: center;
            max-width: 600px;
            margin-left: auto;
            margin-right: auto;
            margin-bottom: 3rem;
        }
        
        /* Step 1: Category Selection */
        .step {
            margin-bottom: 2rem;
        }
        .step-title {
            font-size: 1.5rem;
            margin-bottom: 1.5rem;
            color: #333;
            text-align: center;
        }
        .categories-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }
        .category-card {
            background: white;
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            transition: transform 0.2s, box-shadow 0.2s;
            cursor: pointer;
            border: 2px solid transparent;
            text-align: center;
        }
        .category-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(0,0,0,0.15);
            border-color: #007cba;
        }
        .category-card.selected {
            border-color: #007cba;
            background-color: #f0f8ff;
        }
        .category-icon {
            font-size: 3rem;
            margin-bottom: 1rem;
            color: #007cba;
        }
        .category-name {
            font-size: 1.25rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
            color: #333;
        }
        .category-description {
            color: #666;
            font-size: 0.9rem;
        }
        
        /* Step 2: Package Selection */
        .packages-section {
            display: none;
            margin-top: 3rem;
        }
        .packages-section.active {
            display: block;
        }
        .packages-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }
        .package-card {
            background: white;
            border-radius: 12px;
            padding: 1.5rem;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            transition: transform 0.2s, box-shadow 0.2s;
            position: relative;
            border: 2px solid transparent;
            cursor: pointer;
        }
        .package-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(0,0,0,0.15);
            border-color: #007cba;
        }
        .package-card.selected {
            border-color: #007cba;
            background-color: #f0f8ff;
        }
        .package-card.popular {
            border-color: #ff6b35;
        }
        .popular-badge {
            position: absolute;
            top: -10px;
            right: 15px;
            background: #ff6b35;
            color: white;
            padding: 0.25rem 0.75rem;
            border-radius: 12px;
            font-size: 0.8rem;
            font-weight: 600;
        }
        .package-name {
            font-size: 1.25rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
            color: #333;
        }
        .package-description {
            color: #666;
            margin-bottom: 1rem;
            font-size: 0.9rem;
        }
        .package-details {
            margin-bottom: 1.5rem;
        }
        .package-price {
            font-size: 1.8rem;
            font-weight: 700;
            color: #007cba;
            margin-bottom: 0.5rem;
        }
        .package-count {
            font-size: 1rem;
            color: #666;
            margin-bottom: 0.5rem;
        }
        .price-per-session {
            font-size: 0.9rem;
            color: #333;
            font-weight: 500;
        }
        .savings {
            color: #27ae60;
            font-weight: 600;
            font-size: 0.9rem;
            margin-top: 0.5rem;
        }
        
        /* Purchase Section */
        .purchase-section {
            display: none;
            margin-top: 3rem;
            padding: 2rem;
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }
        .purchase-section.active {
            display: block;
        }
        .purchase-title {
            font-size: 1.5rem;
            margin-bottom: 1rem;
            color: #333;
            text-align: center;
        }
        .selected-details {
            background: #f0f8ff;
            padding: 1.5rem;
            border-radius: 8px;
            margin-bottom: 2rem;
            text-align: center;
        }
        .selected-details h3 {
            font-size: 1.25rem;
            margin-bottom: 0.5rem;
            color: #007cba;
        }
        .selected-details .price {
            font-size: 2rem;
            font-weight: 700;
            color: #007cba;
            margin: 1rem 0;
        }
        .purchase-btn {
            width: 100%;
            padding: 1rem 2rem;
            background-color: #007cba;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 1.1rem;
            font-weight: 600;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        .purchase-btn:hover {
            background-color: #005a87;
        }
        
        @media (max-width: 768px) {
            .categories-grid, .packages-grid {
                grid-template-columns: 1fr;
            }
            .nav-list {
                flex-direction: column;
            }
            .nav-item {
                border-right: none;
                border-bottom: 1px solid #e0e0e0;
            }
            .nav-item:last-child {
                border-bottom: none;
            }
        }
    </style>
</head>
<body>
    <header class="header">
        <h1>Kjernekraft - Elev Dashboard</h1>
    </header>
    
    <nav class="nav">
        <ul class="nav-list">
            <li class="nav-item">
                <a href="/elev/hjem" class="nav-link">Hjem</a>
            </li>
            <li class="nav-item">
                <a href="/elev/timeplan" class="nav-link">Timeplan</a>
            </li>
            <li class="nav-item">
                <a href="/elev/klippekort" class="nav-link active">Klippekort</a>
            </li>
            <li class="nav-item">
                <a href="/elev/medlemskap" class="nav-link">Medlemskap</a>
            </li>
            <li class="nav-item">
                <a href="/elev/min-profil" class="nav-link">Min profil</a>
            </li>
        </ul>
    </nav>
    
    <main class="main">
        <h1 class="page-title">Klippekort</h1>
        <p class="page-description">
            Velg type trening og antall klipp. Jo flere klipp du kjøper, desto mindre blir prisen per økt.
        </p>
        
        <!-- Step 1: Category Selection -->
        <div class="step" id="step1">
            <h2 class="step-title">Steg 1: Velg type trening</h2>
            <div class="categories-grid">
                <div class="category-card" data-category="gruppetimer-sal" onclick="selectCategory('gruppetimer-sal')">
                    <div class="category-icon">🧘‍♀️</div>
                    <h3 class="category-name">Gruppetimer Sal</h3>
                    <p class="category-description">Yoga, Pilates Mat, og andre sal-baserte gruppetimer</p>
                </div>
                
                <div class="category-card" data-category="reformer-apparatus" onclick="selectCategory('reformer-apparatus')">
                    <div class="category-icon">💪</div>
                    <h3 class="category-name">Reformer/Apparatus</h3>
                    <p class="category-description">Spesialisert Pilates utstyr og apparatus-trening</p>
                </div>
                
                <div class="category-card" data-category="self-practice" onclick="selectCategory('self-practice')">
                    <div class="category-icon">🏃‍♂️</div>
                    <h3 class="category-name">Self Practice Pilates Apparatus</h3>
                    <p class="category-description">Selvstendig trening på Pilates apparatus</p>
                </div>
                
                <div class="category-card" data-category="online-gruppetimer" onclick="selectCategory('online-gruppetimer')">
                    <div class="category-icon">💻</div>
                    <h3 class="category-name">Online Gruppetimer</h3>
                    <p class="category-description">Live streaming og on-demand klasser</p>
                </div>
                
                <div class="category-card" data-category="personlig-trening" onclick="selectCategory('personlig-trening')">
                    <div class="category-icon">👨‍💼</div>
                    <h3 class="category-name">Personlig Trening</h3>
                    <p class="category-description">One-on-one trening med personlig trener</p>
                </div>
                
                <div class="category-card" data-category="stressmestring" onclick="selectCategory('stressmestring')">
                    <div class="category-icon">🧠</div>
                    <h3 class="category-name">Stressmestring</h3>
                    <p class="category-description">Avslapning, meditasjon og stresshåndtering</p>
                </div>
            </div>
        </div>
        
        <!-- Step 2: Package Selection for each category -->
        <div class="packages-section" id="packages-gruppetimer-sal">
            <h2 class="step-title">Steg 2: Velg antall klipp for Gruppetimer Sal</h2>
            <div class="packages-grid">
                <div class="package-card" data-package="gruppetimer-5" onclick="selectPackage('gruppetimer-sal', 5, 1200, 'Perfekt for å prøve ut')">
                    <h3 class="package-name">5 klipp</h3>
                    <p class="package-description">Perfekt for å prøve ut</p>
                    <div class="package-details">
                        <div class="package-price">1200 kr</div>
                        <div class="package-count">5 klipp</div>
                        <div class="price-per-session">240 kr per økt</div>
                    </div>
                </div>
                
                <div class="package-card popular" data-package="gruppetimer-10" onclick="selectPackage('gruppetimer-sal', 10, 2200, 'Mest populære pakke')">
                    <div class="popular-badge">Mest populær</div>
                    <h3 class="package-name">10 klipp</h3>
                    <p class="package-description">Mest populære pakke</p>
                    <div class="package-details">
                        <div class="package-price">2200 kr</div>
                        <div class="package-count">10 klipp</div>
                        <div class="price-per-session">220 kr per økt</div>
                        <div class="savings">Spar 200 kr!</div>
                    </div>
                </div>
                
                <div class="package-card" data-package="gruppetimer-20" onclick="selectPackage('gruppetimer-sal', 20, 4000, 'Beste verdi')">
                    <h3 class="package-name">20 klipp</h3>
                    <p class="package-description">Beste verdi</p>
                    <div class="package-details">
                        <div class="package-price">4000 kr</div>
                        <div class="package-count">20 klipp</div>
                        <div class="price-per-session">200 kr per økt</div>
                        <div class="savings">Spar 800 kr!</div>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Repeat for other categories with different pricing -->
        <div class="packages-section" id="packages-reformer-apparatus">
            <h2 class="step-title">Steg 2: Velg antall klipp for Reformer/Apparatus</h2>
            <div class="packages-grid">
                <div class="package-card" data-package="reformer-5" onclick="selectPackage('reformer-apparatus', 5, 2000, 'Prøv Reformer')">
                    <h3 class="package-name">5 klipp</h3>
                    <p class="package-description">Prøv Reformer</p>
                    <div class="package-details">
                        <div class="package-price">2000 kr</div>
                        <div class="package-count">5 klipp</div>
                        <div class="price-per-session">400 kr per økt</div>
                    </div>
                </div>
                
                <div class="package-card popular" data-package="reformer-10" onclick="selectPackage('reformer-apparatus', 10, 3750, 'Populær Reformer pakke')">
                    <div class="popular-badge">Mest populær</div>
                    <h3 class="package-name">10 klipp</h3>
                    <p class="package-description">Populær Reformer pakke</p>
                    <div class="package-details">
                        <div class="package-price">3750 kr</div>
                        <div class="package-count">10 klipp</div>
                        <div class="price-per-session">375 kr per økt</div>
                        <div class="savings">Spar 250 kr!</div>
                    </div>
                </div>
                
                <div class="package-card" data-package="reformer-20" onclick="selectPackage('reformer-apparatus', 20, 7000, 'Best verdi for Reformer')">
                    <h3 class="package-name">20 klipp</h3>
                    <p class="package-description">Best verdi for Reformer</p>
                    <div class="package-details">
                        <div class="package-price">7000 kr</div>
                        <div class="package-count">20 klipp</div>
                        <div class="price-per-session">350 kr per økt</div>
                        <div class="savings">Spar 1000 kr!</div>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="packages-section" id="packages-self-practice">
            <h2 class="step-title">Steg 2: Velg antall klipp for Self Practice</h2>
            <div class="packages-grid">
                <div class="package-card" data-package="self-practice-5" onclick="selectPackage('self-practice', 5, 1500, 'Prøv self practice')">
                    <h3 class="package-name">5 klipp</h3>
                    <p class="package-description">Prøv self practice</p>
                    <div class="package-details">
                        <div class="package-price">1500 kr</div>
                        <div class="package-count">5 klipp</div>
                        <div class="price-per-session">300 kr per økt</div>
                    </div>
                </div>
                
                <div class="package-card popular" data-package="self-practice-10" onclick="selectPackage('self-practice', 10, 2800, 'Populær self practice')">
                    <div class="popular-badge">Mest populær</div>
                    <h3 class="package-name">10 klipp</h3>
                    <p class="package-description">Populær self practice</p>
                    <div class="package-details">
                        <div class="package-price">2800 kr</div>
                        <div class="package-count">10 klipp</div>
                        <div class="price-per-session">280 kr per økt</div>
                        <div class="savings">Spar 200 kr!</div>
                    </div>
                </div>
                
                <div class="package-card" data-package="self-practice-20" onclick="selectPackage('self-practice', 20, 5200, 'Best verdi for self practice')">
                    <h3 class="package-name">20 klipp</h3>
                    <p class="package-description">Best verdi for self practice</p>
                    <div class="package-details">
                        <div class="package-price">5200 kr</div>
                        <div class="package-count">20 klipp</div>
                        <div class="price-per-session">260 kr per økt</div>
                        <div class="savings">Spar 800 kr!</div>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="packages-section" id="packages-online-gruppetimer">
            <h2 class="step-title">Steg 2: Velg antall klipp for Online Gruppetimer</h2>
            <div class="packages-grid">
                <div class="package-card" data-package="online-5" onclick="selectPackage('online-gruppetimer', 5, 800, 'Prøv online')">
                    <h3 class="package-name">5 klipp</h3>
                    <p class="package-description">Prøv online</p>
                    <div class="package-details">
                        <div class="package-price">800 kr</div>
                        <div class="package-count">5 klipp</div>
                        <div class="price-per-session">160 kr per økt</div>
                    </div>
                </div>
                
                <div class="package-card popular" data-package="online-10" onclick="selectPackage('online-gruppetimer', 10, 1400, 'Populær online pakke')">
                    <div class="popular-badge">Mest populær</div>
                    <h3 class="package-name">10 klipp</h3>
                    <p class="package-description">Populær online pakke</p>
                    <div class="package-details">
                        <div class="package-price">1400 kr</div>
                        <div class="package-count">10 klipp</div>
                        <div class="price-per-session">140 kr per økt</div>
                        <div class="savings">Spar 200 kr!</div>
                    </div>
                </div>
                
                <div class="package-card" data-package="online-20" onclick="selectPackage('online-gruppetimer', 20, 2400, 'Best verdi for online')">
                    <h3 class="package-name">20 klipp</h3>
                    <p class="package-description">Best verdi for online</p>
                    <div class="package-details">
                        <div class="package-price">2400 kr</div>
                        <div class="package-count">20 klipp</div>
                        <div class="price-per-session">120 kr per økt</div>
                        <div class="savings">Spar 800 kr!</div>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="packages-section" id="packages-personlig-trening">
            <h2 class="step-title">Steg 2: Velg antall klipp for Personlig Trening</h2>
            <div class="packages-grid">
                <div class="package-card" data-package="pt-5" onclick="selectPackage('personlig-trening', 5, 3000, 'Perfekt for å komme i gang')">
                    <h3 class="package-name">5 klipp</h3>
                    <p class="package-description">Perfekt for å komme i gang</p>
                    <div class="package-details">
                        <div class="package-price">3000 kr</div>
                        <div class="package-count">5 klipp</div>
                        <div class="price-per-session">600 kr per økt</div>
                    </div>
                </div>
                
                <div class="package-card popular" data-package="pt-10" onclick="selectPackage('personlig-trening', 10, 5500, 'Mest populære pakke')">
                    <div class="popular-badge">Mest populær</div>
                    <h3 class="package-name">10 klipp</h3>
                    <p class="package-description">Mest populære pakke</p>
                    <div class="package-details">
                        <div class="package-price">5500 kr</div>
                        <div class="package-count">10 klipp</div>
                        <div class="price-per-session">550 kr per økt</div>
                        <div class="savings">Spar 500 kr!</div>
                    </div>
                </div>
                
                <div class="package-card" data-package="pt-20" onclick="selectPackage('personlig-trening', 20, 10000, 'Best value for money')">
                    <h3 class="package-name">20 klipp</h3>
                    <p class="package-description">Best value for money</p>
                    <div class="package-details">
                        <div class="package-price">10000 kr</div>
                        <div class="package-count">20 klipp</div>
                        <div class="price-per-session">500 kr per økt</div>
                        <div class="savings">Spar 2000 kr!</div>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="packages-section" id="packages-stressmestring">
            <h2 class="step-title">Steg 2: Velg antall klipp for Stressmestring</h2>
            <div class="packages-grid">
                <div class="package-card" data-package="stress-5" onclick="selectPackage('stressmestring', 5, 1000, 'Prøv stressmestring')">
                    <h3 class="package-name">5 klipp</h3>
                    <p class="package-description">Prøv stressmestring</p>
                    <div class="package-details">
                        <div class="package-price">1000 kr</div>
                        <div class="package-count">5 klipp</div>
                        <div class="price-per-session">200 kr per økt</div>
                    </div>
                </div>
                
                <div class="package-card popular" data-package="stress-10" onclick="selectPackage('stressmestring', 10, 1800, 'Populær stressmestring')">
                    <div class="popular-badge">Mest populær</div>
                    <h3 class="package-name">10 klipp</h3>
                    <p class="package-description">Populær stressmestring</p>
                    <div class="package-details">
                        <div class="package-price">1800 kr</div>
                        <div class="package-count">10 klipp</div>
                        <div class="price-per-session">180 kr per økt</div>
                        <div class="savings">Spar 200 kr!</div>
                    </div>
                </div>
                
                <div class="package-card" data-package="stress-20" onclick="selectPackage('stressmestring', 20, 3200, 'Best verdi for stressmestring')">
                    <h3 class="package-name">20 klipp</h3>
                    <p class="package-description">Best verdi for stressmestring</p>
                    <div class="package-details">
                        <div class="package-price">3200 kr</div>
                        <div class="package-count">20 klipp</div>
                        <div class="price-per-session">160 kr per økt</div>
                        <div class="savings">Spar 800 kr!</div>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Purchase Section -->
        <div class="purchase-section" id="purchase-section">
            <h2 class="purchase-title">Bekreft ditt valg</h2>
            <div class="selected-details">
                <h3 id="selected-package-name"></h3>
                <p id="selected-package-description"></p>
                <div class="price" id="selected-package-price"></div>
                <p id="selected-package-details"></p>
            </div>
            <button class="purchase-btn" onclick="purchasePackage()">Kjøp klippekort</button>
        </div>
    </main>

    <script>
        let selectedCategory = null;
        let selectedPackage = null;
        
        function selectCategory(category) {
            // Clear previous selections
            document.querySelectorAll('.category-card').forEach(card => {
                card.classList.remove('selected');
            });
            document.querySelectorAll('.packages-section').forEach(section => {
                section.classList.remove('active');
            });
            document.getElementById('purchase-section').classList.remove('active');
            
            // Select new category
            document.querySelector('.category-card[data-category="' + category + '"]').classList.add('selected');
            document.getElementById('packages-' + category).classList.add('active');
            selectedCategory = category;
            selectedPackage = null;
            
            // Scroll to packages
            document.getElementById('packages-' + category).scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
        
        function selectPackage(category, clips, price, description) {
            // Clear previous package selections
            document.querySelectorAll('.package-card').forEach(card => {
                card.classList.remove('selected');
            });
            
            // Select new package
            event.currentTarget.classList.add('selected');
            
            selectedPackage = {
                category: category,
                clips: clips,
                price: price,
                description: description,
                pricePerSession: price / clips
            };
            
            // Update purchase section
            const categoryNames = {
                'gruppetimer-sal': 'Gruppetimer Sal',
                'reformer-apparatus': 'Reformer/Apparatus',
                'self-practice': 'Self Practice Pilates Apparatus',
                'online-gruppetimer': 'Online Gruppetimer',
                'personlig-trening': 'Personlig Trening',
                'stressmestring': 'Stressmestring'
            };
            
            document.getElementById('selected-package-name').textContent = clips + ' klipp ' + categoryNames[category];
            document.getElementById('selected-package-description').textContent = description;
            document.getElementById('selected-package-price').textContent = price + ' kr';
            document.getElementById('selected-package-details').textContent = clips + ' klipp • ' + Math.round(selectedPackage.pricePerSession) + ' kr per økt';
            
            // Show purchase section
            document.getElementById('purchase-section').classList.add('active');
            document.getElementById('purchase-section').scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
        
        function purchasePackage() {
            if (!selectedPackage) {
                alert('Vennligst velg en pakke først');
                return;
            }
            
            // TODO: Implement actual purchase functionality
            alert('Kjøp av ' + selectedPackage.clips + ' klipp for ' + selectedPackage.price + ' kr - kommer snart!');
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}

// MembershipSelectorHandler serves the interactive membership selector page
func MembershipSelectorHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Finn ditt medlemskap - Kjernekraft</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f5f5f5;
            color: #333;
        }
        .header {
            background-color: #007cba;
            color: white;
            padding: 1rem 2rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header h1 {
            font-size: 1.5rem;
            font-weight: 600;
        }
        .nav {
            background-color: white;
            border-bottom: 1px solid #e0e0e0;
            padding: 0;
        }
        .nav-list {
            display: flex;
            list-style: none;
            max-width: 1200px;
            margin: 0 auto;
        }
        .nav-item {
            border-right: 1px solid #e0e0e0;
        }
        .nav-item:last-child {
            border-right: none;
        }
        .nav-link {
            display: block;
            padding: 1rem 2rem;
            text-decoration: none;
            color: #333;
            font-weight: 500;
            transition: background-color 0.2s;
        }
        .nav-link:hover, .nav-link.active {
            background-color: #f0f8ff;
            color: #007cba;
        }
        .main {
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
        }
        .page-title {
            font-size: 2rem;
            margin-bottom: 1rem;
            color: #333;
            text-align: center;
        }
        .page-description {
            font-size: 1.1rem;
            color: #666;
            margin-bottom: 3rem;
            text-align: center;
            max-width: 600px;
            margin-left: auto;
            margin-right: auto;
            margin-bottom: 3rem;
        }
        .selector-container {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 3rem;
            align-items: start;
        }
        .question-form {
            background: white;
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }
        .form-title {
            font-size: 1.25rem;
            margin-bottom: 1.5rem;
            color: #333;
        }
        .question-group {
            margin-bottom: 1.5rem;
        }
        .question-label {
            display: block;
            margin-bottom: 0.5rem;
            font-weight: 600;
            color: #333;
        }
        .question-options {
            display: grid;
            gap: 0.5rem;
        }
        .option-label {
            display: flex;
            align-items: center;
            padding: 0.75rem;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.2s;
        }
        .option-label:hover {
            border-color: #007cba;
            background-color: #f8f9fa;
        }
        .option-label input[type="radio"] {
            margin-right: 0.75rem;
        }
        .option-label input[type="radio"]:checked + span {
            font-weight: 600;
        }
        .option-label:has(input[type="radio"]:checked) {
            border-color: #007cba;
            background-color: #e8f4fd;
        }
        .checkout-container {
            background: white;
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            position: sticky;
            top: 2rem;
        }
        .checkout-title {
            font-size: 1.25rem;
            margin-bottom: 1.5rem;
            color: #333;
            text-align: center;
        }
        .price-display {
            text-align: center;
            margin-bottom: 2rem;
            padding: 1.5rem;
            background: #f8f9fa;
            border-radius: 8px;
        }
        .current-price {
            font-size: 2.5rem;
            font-weight: 700;
            color: #007cba;
            margin-bottom: 0.5rem;
        }
        .original-price {
            font-size: 1.1rem;
            color: #666;
            text-decoration: line-through;
            margin-bottom: 0.5rem;
        }
        .discount-badge {
            background: #28a745;
            color: white;
            padding: 0.25rem 0.75rem;
            border-radius: 12px;
            font-size: 0.8rem;
            font-weight: 600;
            display: inline-block;
        }
        .membership-details {
            margin-bottom: 2rem;
            padding: 1rem;
            background: #fff;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
        }
        .detail-row {
            display: flex;
            justify-content: space-between;
            margin-bottom: 0.5rem;
        }
        .detail-row:last-child {
            margin-bottom: 0;
            font-weight: 600;
            padding-top: 0.5rem;
            border-top: 1px solid #e0e0e0;
        }
        .checkout-btn {
            width: 100%;
            background: #007cba;
            color: white;
            border: none;
            padding: 1rem 2rem;
            border-radius: 8px;
            font-size: 1.1rem;
            font-weight: 600;
            cursor: pointer;
            transition: background-color 0.2s;
            margin-bottom: 1rem;
        }
        .checkout-btn:hover {
            background: #005a87;
        }
        .checkout-btn:disabled {
            background: #adb5bd;
            cursor: not-allowed;
        }
        .special-offer-notice {
            background: linear-gradient(135deg, #ff6b35, #f7931e);
            color: white;
            padding: 1rem;
            border-radius: 8px;
            margin-bottom: 1rem;
            text-align: center;
            font-weight: 600;
        }
        @media (max-width: 768px) {
            .selector-container {
                grid-template-columns: 1fr;
                gap: 2rem;
            }
            .nav-list {
                flex-direction: column;
                gap: 0.5rem;
            }
        }
    </style>
</head>
<body>
    <header class="header">
        <h1>Kjernekraft - Finn ditt medlemskap</h1>
    </header>
    
    <nav class="nav">
        <ul class="nav-list">
            <li class="nav-item">
                <a href="/elev/hjem" class="nav-link">Hjem</a>
            </li>
            <li class="nav-item">
                <a href="/elev/timeplan" class="nav-link">Timeplan</a>
            </li>
            <li class="nav-item">
                <a href="/elev/klippekort" class="nav-link">Klippekort</a>
            </li>
            <li class="nav-item">
                <a href="/elev/medlemskap" class="nav-link active">Medlemskap</a>
            </li>
            <li class="nav-item">
                <a href="/elev/min-profil" class="nav-link">Min profil</a>
            </li>
        </ul>
    </nav>
    
    <main class="main">
        <h1 class="page-title">Finn ditt perfekte medlemskap</h1>
        <p class="page-description">
            Svar på noen enkle spørsmål så viser vi deg medlemskapet som passer best for deg.
        </p>
        
        <div class="selector-container">
            <form class="question-form" 
                  action="/medlemskap/recommendations" 
                  method="post">
                
                <h2 class="form-title">Fortell oss om deg</h2>
                
                <div class="question-group">
                    <label class="question-label">Er du student eller senior?</label>
                    <div class="question-options">
                        <label class="option-label">
                            <input type="radio" name="is_student_senior" value="true">
                            <span>Ja, jeg er student eller 67+ år</span>
                        </label>
                        <label class="option-label">
                            <input type="radio" name="is_student_senior" value="false">
                            <span>Nei</span>
                        </label>
                    </div>
                </div>
                
                <div class="question-group">
                    <label class="question-label">Hvor lenge ønsker du å binde deg?</label>
                    <div class="question-options">
                        <label class="option-label">
                            <input type="radio" name="commitment" value="12">
                            <span>12 måneder (best pris)</span>
                        </label>
                        <label class="option-label">
                            <input type="radio" name="commitment" value="6">
                            <span>6 måneder</span>
                        </label>
                        <label class="option-label">
                            <input type="radio" name="commitment" value="0">
                            <span>Ingen binding (mest fleksibelt)</span>
                        </label>
                        <label class="option-label">
                            <input type="radio" name="commitment" value="trial">
                            <span>Jeg vil bare prøve</span>
                        </label>
                    </div>
                </div>
                
                <div class="question-group">
                    <label class="question-label">Når vil du starte?</label>
                    <div class="question-options">
                        <label class="option-label">
                            <input type="radio" name="start_time" value="now" checked>
                            <span>Så snart som mulig</span>
                        </label>
                        <label class="option-label">
                            <input type="radio" name="start_time" value="august">
                            <span>I august (Høsttilbud!)</span>
                        </label>
                    </div>
                </div>
            </form>
            
            <div class="checkout-container">
                <h2 class="checkout-title">Ditt medlemskap</h2>
                
                <div id="special-offer" class="special-offer-notice" style="display: none;">
                    🍂 Høsttilbud: 12-måneders pris med kun 4 måneders binding!
                </div>
                
                <div class="price-display">
                    <div id="current-price" class="current-price">1490 kr/mnd</div>
                    <div id="original-price" class="original-price" style="display: none;"></div>
                    <div id="discount-badge" class="discount-badge" style="display: none;">20% rabatt</div>
                </div>
                
                <div class="membership-details">
                    <div class="detail-row">
                        <span>Medlemskap:</span>
                        <span id="membership-type">Ingen binding</span>
                    </div>
                    <div class="detail-row">
                        <span>Binding:</span>
                        <span id="binding-period">Ingen</span>
                    </div>
                    <div class="detail-row">
                        <span>Student/Senior rabatt:</span>
                        <span id="discount-status">Nei</span>
                    </div>
                    <div class="detail-row">
                        <span>Total pris per måned:</span>
                        <span id="total-price">1490 kr</span>
                    </div>
                </div>
                
                <button class="checkout-btn" onclick="proceedToCheckout()">
                    Fortsett til betaling
                </button>
                
                <p style="font-size: 0.9rem; color: #666; text-align: center;">
                    Ingen skjulte kostnader. Avbryt når som helst.
                </p>
            </div>
        </div>
    </main>

    <script>
        // Default pricing data (most expensive, no commitment option)
        const pricingData = {
            'no_binding': { regular: 149000, student: 119200 }, // 1490kr, 20% discount = 1192kr
            '6_months': { regular: 129000, student: 103200 },   // 1290kr, 20% discount = 1032kr  
            '12_months': { regular: 104000, student: 83200 },   // 1040kr, 20% discount = 832kr
            'trial': { regular: 39900, student: 39900 }         // 399kr (trial price same for all)
        };

        let currentSelection = {
            isStudentSenior: false,
            commitment: '0', // Default to no binding (most expensive)
            startTime: 'now'
        };

        function updatePricing() {
            const isStudentSenior = currentSelection.isStudentSenior;
            const commitment = currentSelection.commitment;
            const startTime = currentSelection.startTime;

            let priceKey = 'no_binding';
            let membershipType = 'Ingen binding';
            let bindingText = 'Ingen';

            if (commitment === '12') {
                priceKey = '12_months';
                membershipType = '12-måneder';
                bindingText = '12 måneder';
            } else if (commitment === '6') {
                priceKey = '6_months';
                membershipType = '6-måneder';
                bindingText = '6 måneder';
            } else if (commitment === 'trial') {
                priceKey = 'trial';
                membershipType = 'Prøvemedlemskap';
                bindingText = '2 uker';
            }

            const pricing = pricingData[priceKey];
            const originalPrice = pricing.regular;
            const currentPrice = isStudentSenior ? pricing.student : pricing.regular;

            // Update UI elements
            document.getElementById('current-price').textContent = Math.round(currentPrice / 100) + ' kr/mnd';
            document.getElementById('membership-type').textContent = membershipType;
            document.getElementById('binding-period').textContent = bindingText;
            document.getElementById('discount-status').textContent = isStudentSenior ? 'Ja (20%)' : 'Nei';
            document.getElementById('total-price').textContent = Math.round(currentPrice / 100) + ' kr';

            // Show/hide discount elements
            if (isStudentSenior && priceKey !== 'trial') {
                document.getElementById('original-price').textContent = Math.round(originalPrice / 100) + ' kr/mnd';
                document.getElementById('original-price').style.display = 'block';
                document.getElementById('discount-badge').style.display = 'inline-block';
            } else {
                document.getElementById('original-price').style.display = 'none';
                document.getElementById('discount-badge').style.display = 'none';
            }

            // Show/hide special offer notice
            if (startTime === 'august') {
                document.getElementById('special-offer').style.display = 'block';
            } else {
                document.getElementById('special-offer').style.display = 'none';
            }
        }

        function proceedToCheckout() {
            alert('Checkout funksjonalitet kommer snart! Valgt medlemskap: ' + 
                  document.getElementById('membership-type').textContent + 
                  ' til ' + document.getElementById('total-price').textContent);
        }

        // Event listeners for form changes
        document.addEventListener('DOMContentLoaded', function() {
            const form = document.querySelector('.question-form');
            
            form.addEventListener('change', function(e) {
                if (e.target.name === 'is_student_senior') {
                    currentSelection.isStudentSenior = e.target.value === 'true';
                } else if (e.target.name === 'commitment') {
                    currentSelection.commitment = e.target.value;
                } else if (e.target.name === 'start_time') {
                    currentSelection.startTime = e.target.value;
                }
                
                updatePricing();
            });

            // Initialize with default pricing
            updatePricing();
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}

// MembershipRecommendationsHandler provides endpoint for membership filtering
func MembershipRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	isStudentSenior := r.FormValue("is_student_senior") == "true"
	commitment := r.FormValue("commitment")
	startTime := r.FormValue("start_time")

	// Get all memberships
	allMemberships, err := DB.GetAllMemberships()
	if err != nil {
		http.Error(w, "Could not fetch memberships", http.StatusInternalServerError)
		return
	}

	// Filter memberships based on criteria
	var recommendations []models.Membership
	for _, membership := range allMemberships {
		// Check student/senior eligibility
		if isStudentSenior != membership.IsStudentSenior {
			continue
		}

		// Check commitment preferences
		if commitment == "trial" {
			// Show trial options (2-week trial, monthly pass)
			if membership.ID == 7 || membership.ID == 8 {
				recommendations = append(recommendations, membership)
			}
		} else if commitment != "" {
			commitmentMonths, _ := strconv.Atoi(commitment)
			if membership.CommitmentMonths == commitmentMonths {
				recommendations = append(recommendations, membership)
			}
		}

		// Special handling for autumn offer
		if startTime == "august" && membership.IsSpecialOffer {
			recommendations = append(recommendations, membership)
		}
	}

	// If no specific matches, show some default options
	if len(recommendations) == 0 && commitment != "" {
		for _, membership := range allMemberships {
			if membership.IsStudentSenior == isStudentSenior && !membership.IsSpecialOffer {
				recommendations = append(recommendations, membership)
			}
		}
	}

	// Check if this is an HTMX request
	isHTMX := r.Header.Get("HX-Request") != ""
	
	if isHTMX {
		// Return HTML fragment for HTMX
		data := struct {
			Recommendations []models.Membership
			ShowAutumnOffer bool
		}{
			Recommendations: recommendations,
			ShowAutumnOffer: startTime == "august",
		}

		tmpl := `{{if .Recommendations}}
<div style="background: white; border-radius: 12px; padding: 1.5rem; box-shadow: 0 4px 12px rgba(0,0,0,0.1);">
    <h3 style="margin-bottom: 1.5rem; color: #333; font-size: 1.25rem;">Våre anbefalinger for deg:</h3>
    
    {{if .ShowAutumnOffer}}
    <div style="background: linear-gradient(135deg, #ff6b35, #f7931e); color: white; padding: 1rem; border-radius: 8px; margin-bottom: 1.5rem; text-align: center;">
        <strong>🍂 Spesielt Høsttilbud!</strong><br>
        Få 12-måneders pris med kun 4 måneders binding
    </div>
    {{end}}
    
    <div style="display: grid; gap: 1rem;">
        {{range .Recommendations}}
        <div style="border: 2px solid {{if .IsSpecialOffer}}#ff6b35{{else}}#e0e0e0{{end}}; border-radius: 8px; padding: 1.5rem; {{if .IsSpecialOffer}}background-color: #fff5f0;{{end}}">
            {{if .IsSpecialOffer}}
            <div style="background: #ff6b35; color: white; padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.8rem; font-weight: 600; display: inline-block; margin-bottom: 0.5rem;">
                Spesialtilbud
            </div>
            {{end}}
            
            <h4 style="font-size: 1.1rem; margin-bottom: 0.5rem; color: #333;">{{.Name}}</h4>
            <p style="color: #666; margin-bottom: 1rem; font-size: 0.9rem;">{{.Description}}</p>
            
            <div style="display: flex; justify-content: space-between; align-items: center;">
                <div>
                    <div style="font-size: 1.5rem; font-weight: 700; color: #007cba;">{{printf "%.0f" (divf .Price 100)}} kr/mnd</div>
                    {{if gt .CommitmentMonths 0}}
                    <div style="font-size: 0.8rem; color: #666;">{{.CommitmentMonths}} måneders binding</div>
                    {{else}}
                    <div style="font-size: 0.8rem; color: #666;">Ingen binding</div>
                    {{end}}
                </div>
                <button style="background: #007cba; color: white; border: none; padding: 0.75rem 1.5rem; border-radius: 6px; cursor: pointer; font-weight: 600;">
                    Velg dette
                </button>
            </div>
        </div>
        {{end}}
    </div>
</div>
{{else}}
<div style="background: white; border-radius: 12px; padding: 2rem; box-shadow: 0 4px 12px rgba(0,0,0,0.1); text-align: center; color: #666;">
    Velg flere alternativer for å se anbefalinger
</div>
{{end}}`

		// Parse template with custom functions
		tmplFuncs := template.FuncMap{
			"divf": func(a, b interface{}) float64 {
				var aFloat, bFloat float64
				
				switch v := a.(type) {
				case int:
					aFloat = float64(v)
				case float64:
					aFloat = v
				default:
					return 0
				}
				
				switch v := b.(type) {
				case int:
					bFloat = float64(v)
				case float64:
					bFloat = v
				default:
					return 0
				}
				
				if bFloat == 0 {
					return 0
				}
				return aFloat / bFloat
			},
		}

		t, err := template.New("recommendations").Funcs(tmplFuncs).Parse(tmpl)
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := t.Execute(w, data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
	} else {
		// Return full page for regular form submission
		data := struct {
			Recommendations []models.Membership
			ShowAutumnOffer bool
			IsStudentSenior bool
			Commitment      string
			StartTime       string
		}{
			Recommendations: recommendations,
			ShowAutumnOffer: startTime == "august",
			IsStudentSenior: isStudentSenior,
			Commitment:      commitment,
			StartTime:       startTime,
		}

		tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Medlemskapsanbefalinger - Kjernekraft</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5; color: #333; }
        .header { background-color: #007cba; color: white; padding: 1rem 2rem; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header h1 { font-size: 1.5rem; font-weight: 600; }
        .nav { background-color: white; border-bottom: 1px solid #e0e0e0; padding: 0; }
        .nav-list { display: flex; list-style: none; max-width: 1200px; margin: 0 auto; }
        .nav-item { border-right: 1px solid #e0e0e0; }
        .nav-item:last-child { border-right: none; }
        .nav-link { display: block; padding: 1rem 2rem; text-decoration: none; color: #333; font-weight: 500; transition: background-color 0.2s; }
        .nav-link:hover, .nav-link.active { background-color: #f0f8ff; color: #007cba; }
        .main { max-width: 800px; margin: 0 auto; padding: 2rem; }
        .page-title { font-size: 2rem; margin-bottom: 2rem; color: #333; text-align: center; }
        .recommendations { display: grid; gap: 1.5rem; }
        .recommendation-card { background: white; border-radius: 12px; padding: 2rem; box-shadow: 0 4px 12px rgba(0,0,0,0.1); border: 2px solid #e0e0e0; }
        .recommendation-card.special { border-color: #ff6b35; background: linear-gradient(135deg, #fff5f0, #ffffff); }
        .special-badge { background: #ff6b35; color: white; padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.8rem; font-weight: 600; display: inline-block; margin-bottom: 1rem; }
        .card-title { font-size: 1.5rem; font-weight: 600; color: #333; margin-bottom: 0.5rem; }
        .card-description { color: #666; margin-bottom: 1.5rem; }
        .price { font-size: 2rem; font-weight: 700; color: #007cba; margin-bottom: 0.5rem; }
        .commitment { color: #666; margin-bottom: 1.5rem; }
        .back-link { display: inline-block; margin-top: 2rem; padding: 0.75rem 1.5rem; background: #6c757d; color: white; text-decoration: none; border-radius: 6px; }
        .back-link:hover { background: #5a6268; }
    </style>
</head>
<body>
    <header class="header"><h1>Kjernekraft - Medlemskapsanbefalinger</h1></header>
    <nav class="nav">
        <ul class="nav-list">
            <li class="nav-item"><a href="/elev/hjem" class="nav-link">Hjem</a></li>
            <li class="nav-item"><a href="/elev/timeplan" class="nav-link">Timeplan</a></li>
            <li class="nav-item"><a href="/elev/klippekort" class="nav-link">Klippekort</a></li>
            <li class="nav-item"><a href="/elev/medlemskap" class="nav-link active">Medlemskap</a></li>
            <li class="nav-item"><a href="/elev/min-profil" class="nav-link">Min profil</a></li>
        </ul>
    </nav>
    
    <main class="main">
        <h1 class="page-title">Våre anbefalinger for deg</h1>
        
        {{if .ShowAutumnOffer}}
        <div style="background: linear-gradient(135deg, #ff6b35, #f7931e); color: white; padding: 1.5rem; border-radius: 12px; margin-bottom: 2rem; text-align: center;">
            <strong style="font-size: 1.2rem;">🍂 Spesielt Høsttilbud!</strong><br>
            Få 12-måneders pris med kun 4 måneders binding
        </div>
        {{end}}
        
        <div class="recommendations">
            {{range .Recommendations}}
            <div class="recommendation-card {{if .IsSpecialOffer}}special{{end}}">
                {{if .IsSpecialOffer}}
                <div class="special-badge">Spesialtilbud</div>
                {{end}}
                
                <h2 class="card-title">{{.Name}}</h2>
                <p class="card-description">{{.Description}}</p>
                
                <div class="price">{{printf "%.0f" (divf .Price 100)}} kr/mnd</div>
                {{if gt .CommitmentMonths 0}}
                <div class="commitment">{{.CommitmentMonths}} måneders binding</div>
                {{else}}
                <div class="commitment">Ingen binding</div>
                {{end}}
            </div>
            {{end}}
        </div>
        
        <a href="/medlemskap" class="back-link">← Tilbake til spørreskjema</a>
    </main>
</body>
</html>`

		// Parse template with custom functions
		tmplFuncs := template.FuncMap{
			"divf": func(a, b interface{}) float64 {
				var aFloat, bFloat float64
				
				switch v := a.(type) {
				case int:
					aFloat = float64(v)
				case float64:
					aFloat = v
				default:
					return 0
				}
				
				switch v := b.(type) {
				case int:
					bFloat = float64(v)
				case float64:
					bFloat = v
				default:
					return 0
				}
				
				if bFloat == 0 {
					return 0
				}
				return aFloat / bFloat
			},
		}

		t, err := template.New("membership-results").Funcs(tmplFuncs).Parse(tmpl)
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := t.Execute(w, data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
	}
}

// MinProfilHandler serves the user profile page
func MinProfilHandler(w http.ResponseWriter, r *http.Request) {
	// For now, use a test user. In a real app, this would come from session/auth
	user := struct {
		ID       int64
		Name     string
		Email    string
		JoinDate string
		Phone    string
	}{
		ID:       1,
		Name:     "Test Bruker",
		Email:    "test@example.com",
		JoinDate: "1. januar 2024",
		Phone:    "+47 123 45 678",
	}

	tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Min profil - Kjernekraft</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f5f5f5;
            color: #333;
        }
        .header {
            background-color: #007cba;
            color: white;
            padding: 1rem 2rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header h1 {
            font-size: 1.5rem;
            font-weight: 600;
        }
        .nav {
            background-color: white;
            border-bottom: 1px solid #e0e0e0;
            padding: 0;
        }
        .nav-list {
            display: flex;
            list-style: none;
            max-width: 1200px;
            margin: 0 auto;
        }
        .nav-item {
            border-right: 1px solid #e0e0e0;
        }
        .nav-item:last-child {
            border-right: none;
        }
        .nav-link {
            display: block;
            padding: 1rem 2rem;
            text-decoration: none;
            color: #333;
            font-weight: 500;
            transition: background-color 0.2s;
        }
        .nav-link:hover, .nav-link.active {
            background-color: #f0f8ff;
            color: #007cba;
        }
        .main-content {
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
        }
        .page-title {
            font-size: 2rem;
            margin-bottom: 2rem;
            color: #333;
        }
        .profile-card {
            background: white;
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            margin-bottom: 2rem;
        }
        .profile-header {
            display: flex;
            align-items: center;
            margin-bottom: 2rem;
            padding-bottom: 1rem;
            border-bottom: 1px solid #e0e0e0;
        }
        .profile-avatar {
            width: 80px;
            height: 80px;
            background: linear-gradient(135deg, #007cba, #4a9fd1);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 2rem;
            color: white;
            font-weight: 600;
            margin-right: 1.5rem;
        }
        .profile-info h2 {
            font-size: 1.5rem;
            color: #333;
            margin-bottom: 0.5rem;
        }
        .profile-info .email {
            color: #666;
            font-size: 1rem;
        }
        .profile-details {
            display: grid;
            gap: 1rem;
        }
        .detail-item {
            display: flex;
            justify-content: space-between;
            padding: 1rem 0;
            border-bottom: 1px solid #f0f0f0;
        }
        .detail-item:last-child {
            border-bottom: none;
        }
        .detail-label {
            font-weight: 600;
            color: #333;
        }
        .detail-value {
            color: #666;
        }
        .edit-profile-btn {
            background: #007cba;
            color: white;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 6px;
            font-weight: 600;
            cursor: pointer;
            transition: background-color 0.2s;
            margin-top: 1rem;
        }
        .edit-profile-btn:hover {
            background: #005a87;
        }
        /* Responsive styles */
        @media (max-width: 767px) {
            .nav-list {
                flex-direction: column;
            }
            .nav-item {
                border-right: none;
                border-bottom: 1px solid #e0e0e0;
            }
            .nav-item:last-child {
                border-bottom: none;
            }
            .main-content {
                padding: 1rem;
            }
        }
        @media (max-width: 768px) {
            .profile-header {
                flex-direction: column;
                text-align: center;
            }
            .profile-avatar {
                margin-right: 0;
                margin-bottom: 1rem;
            }
        }
    </style>
</head>
<body>
    <header class="header">
        <h1>Kjernekraft - Min profil</h1>
    </header>
    
    <nav class="nav">
        <ul class="nav-list">
            <li class="nav-item">
                <a href="/elev/hjem" class="nav-link">Hjem</a>
            </li>
            <li class="nav-item">
                <a href="/elev/timeplan" class="nav-link">Timeplan</a>
            </li>
            <li class="nav-item">
                <a href="/elev/klippekort" class="nav-link">Klippekort</a>
            </li>
            <li class="nav-item">
                <a href="/elev/medlemskap" class="nav-link">Medlemskap</a>
            </li>
            <li class="nav-item">
                <a href="/elev/min-profil" class="nav-link active">Min profil</a>
            </li>
        </ul>
    </nav>

    <main class="main-content">
        <h1 class="page-title">Min profil</h1>
        
        <div class="profile-card">
            <div class="profile-header">
                <div class="profile-avatar">{{substr .Name 0 1}}</div>
                <div class="profile-info">
                    <h2>{{.Name}}</h2>
                    <div class="email">{{.Email}}</div>
                </div>
            </div>
            
            <div class="profile-details">
                <div class="detail-item">
                    <span class="detail-label">Fullt navn:</span>
                    <span class="detail-value">{{.Name}}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">E-post:</span>
                    <span class="detail-value">{{.Email}}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Telefon:</span>
                    <span class="detail-value">{{.Phone}}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Medlem siden:</span>
                    <span class="detail-value">{{.JoinDate}}</span>
                </div>
                <div class="detail-item">
                    <span class="detail-label">Bruker-ID:</span>
                    <span class="detail-value">#{{.ID}}</span>
                </div>
            </div>
            
            <button class="edit-profile-btn" onclick="editProfile()">
                Rediger profil
            </button>
        </div>
    </main>

    <script>
        function editProfile() {
            alert('Redigering av profil kommer snart!');
        }
    </script>
</body>
</html>`

	t, err := template.New("min-profil").Funcs(template.FuncMap{
		"substr": func(s string, start int, length int) string {
			if start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
	}).Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, user)
}

// TestDataPageHandler serves the test data generation page
func TestDataPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Testdata - Kjernekraft</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f5f5f5;
            color: #333;
        }
        .header {
            background-color: #007cba;
            color: white;
            padding: 1rem 2rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header h1 {
            font-size: 1.5rem;
            font-weight: 600;
        }
        .nav {
            background-color: white;
            border-bottom: 1px solid #e0e0e0;
            padding: 0;
        }
        .nav-list {
            display: flex;
            list-style: none;
            max-width: 1200px;
            margin: 0 auto;
        }
        .nav-item {
            border-right: 1px solid #e0e0e0;
        }
        .nav-item:last-child {
            border-right: none;
        }
        .nav-link {
            display: block;
            padding: 1rem 2rem;
            text-decoration: none;
            color: #333;
            font-weight: 500;
            transition: background-color 0.2s;
        }
        .nav-link:hover, .nav-link.active {
            background-color: #f0f8ff;
            color: #007cba;
        }
        .main-content {
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
        }
        .page-title {
            font-size: 2rem;
            margin-bottom: 1rem;
            color: #333;
        }
        .dev-warning {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            color: #856404;
            padding: 1rem;
            border-radius: 8px;
            margin-bottom: 2rem;
        }
        .test-section {
            background: white;
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            margin-bottom: 2rem;
        }
        .section-title {
            font-size: 1.25rem;
            margin-bottom: 1rem;
            color: #333;
        }
        .section-description {
            color: #666;
            margin-bottom: 1.5rem;
            line-height: 1.6;
        }
        .test-btn {
            background: #6c757d;
            color: white;
            border: none;
            padding: 1rem 2rem;
            border-radius: 6px;
            font-weight: 600;
            cursor: pointer;
            transition: background-color 0.2s;
            margin-right: 1rem;
            margin-bottom: 0.5rem;
        }
        .test-btn:hover {
            background: #5a6268;
        }
        .test-btn:disabled {
            background: #adb5bd;
            cursor: not-allowed;
        }
        .test-btn.danger {
            background: #dc3545;
        }
        .test-btn.danger:hover {
            background: #c82333;
        }
        .result-area {
            margin-top: 1rem;
            padding: 1rem;
            background: #f8f9fa;
            border-radius: 6px;
            display: none;
        }
        .result-area.show {
            display: block;
        }
        .result-area.success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .result-area.error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f1aeb5;
        }
    </style>
</head>
<body>
    <header class="header">
        <h1>Kjernekraft - Testdata</h1>
    </header>
    
    <nav class="nav">
        <ul class="nav-list">
            <li class="nav-item">
                <a href="/elev/hjem" class="nav-link">Hjem</a>
            </li>
            <li class="nav-item">
                <a href="/elev/timeplan" class="nav-link">Timeplan</a>
            </li>
            <li class="nav-item">
                <a href="/elev/klippekort" class="nav-link">Klippekort</a>
            </li>
            <li class="nav-item">
                <a href="/elev/medlemskap" class="nav-link">Medlemskap</a>
            </li>
            <li class="nav-item">
                <a href="/elev/min-profil" class="nav-link">Min profil</a>
            </li>
        </ul>
    </nav>

    <main class="main-content">
        <h1 class="page-title">🧪 Testdata generering</h1>
        
        <div class="dev-warning">
            <strong>⚠️ Utviklingsverktøy</strong><br>
            Denne siden er kun tilgjengelig i utviklingsmiljø og vil generere testdata for demonstrasjon.
        </div>
        
        <div class="test-section">
            <h2 class="section-title">Kalenderdata</h2>
            <p class="section-description">
                Generer nye tilfeldige treningsklasser for denne og neste uke. Dette vil erstatte alle eksisterende kalenderoppføringer.
            </p>
            <button class="test-btn" onclick="shuffleEvents()">
                🗓️ Generer kalenderdata
            </button>
            <div id="events-result" class="result-area"></div>
        </div>
        
        <div class="test-section">
            <h2 class="section-title">Medlemskapsdata</h2>
            <p class="section-description">
                Generer nye tilfeldige medlemskapsnavn og priser. Dette vil oppdatere eksisterende medlemskapstyper med nye verdier.
            </p>
            <button class="test-btn" onclick="shuffleMemberships()">
                💳 Generer medlemskapsdata
            </button>
            <div id="memberships-result" class="result-area"></div>
        </div>
        
        <div class="test-section">
            <h2 class="section-title">Brukerdata</h2>
            <p class="section-description">
                Oppdater den innloggede brukerens klippekort med nye tilfeldige verdier. Dette endrer antall gjenværende klipp.
            </p>
            <button class="test-btn" onclick="shuffleUserKlippekort()">
                🎫 Generer brukerklippekort
            </button>
            <div id="user-result" class="result-area"></div>
        </div>
        
        <div class="test-section">
            <h2 class="section-title">Generer alt</h2>
            <p class="section-description">
                Generer alle testdata på en gang. Dette vil oppdatere kalenderen, medlemskap og brukerdata samtidig.
            </p>
            <button class="test-btn danger" onclick="shuffleAll()">
                🎲 Generer alle testdata
            </button>
            <div id="all-result" class="result-area"></div>
        </div>
    </main>

    <script>
        async function makeRequest(endpoint, btnElement, resultElement, successMessage) {
            btnElement.disabled = true;
            btnElement.textContent = '🔄 Genererer...';
            resultElement.className = 'result-area';
            resultElement.style.display = 'none';
            
            try {
                const response = await fetch(endpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                });
                
                if (response.ok) {
                    const data = await response.json();
                    resultElement.className = 'result-area success show';
                    resultElement.textContent = successMessage + (data.message ? ': ' + data.message : '');
                } else {
                    throw new Error('Request failed');
                }
            } catch (error) {
                console.error('Error:', error);
                resultElement.className = 'result-area error show';
                resultElement.textContent = 'Feil ved generering av testdata';
            } finally {
                btnElement.disabled = false;
                btnElement.textContent = btnElement.textContent.replace('🔄 Genererer...', btnElement.getAttribute('data-original-text'));
            }
        }

        function shuffleEvents() {
            const btn = event.target;
            btn.setAttribute('data-original-text', '🗓️ Generer kalenderdata');
            const result = document.getElementById('events-result');
            makeRequest('/api/shuffle-test-data', btn, result, 'Kalenderdata generert');
        }

        function shuffleMemberships() {
            const btn = event.target;
            btn.setAttribute('data-original-text', '💳 Generer medlemskapsdata');
            const result = document.getElementById('memberships-result');
            makeRequest('/api/shuffle-memberships', btn, result, 'Medlemskapsdata generert');
        }

        function shuffleUserKlippekort() {
            const btn = event.target;
            btn.setAttribute('data-original-text', '🎫 Generer brukerklippekort');
            const result = document.getElementById('user-result');
            makeRequest('/api/shuffle-user-klippekort', btn, result, 'Brukerklippekort generert');
        }

        function shuffleAll() {
            const btn = event.target;
            btn.setAttribute('data-original-text', '🎲 Generer alle testdata');
            const result = document.getElementById('all-result');
            makeRequest('/api/shuffle-all-test-data', btn, result, 'Alle testdata generert');
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}