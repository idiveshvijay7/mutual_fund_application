import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import LoginPage from './pages/login';
import HomePage from './pages/homepage';
import HoldingPage from './pages/holding';
import StrategySelectionPage from './pages/selectingStrategy'; // Import the StrategySelectionPage component
import TransitPage from './pages/transit'; // Import the TransitPage component
import SuccessfulPage from './constant/SuccessfulInvestment';
import FailurePage from './constant/FailureInvestment';


function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<LoginPage />} />
        <Route path="/home" element={<HomePage />} />
        <Route path="/holdings" element={<HoldingPage />} />
        <Route path="/select-strategy" element={<StrategySelectionPage />} /> {/* Add a route for StrategySelectionPage */}
        <Route path="/transit/:strategyName" element={<TransitPage />} /> {/* Add a route for TransitPage */}
        <Route path="/investmentSuccessful" element={<SuccessfulPage />} />
        <Route path="/investmentFailure" element={<FailurePage />} />
      </Routes>
    </Router>
  );
}

export default App;
