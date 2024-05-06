// StrategySelectionPage.js
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import strategies from '../constant/strategies'; // Import the strategies data

const StrategySelectionPage = () => {
    const [hoveredIndex, setHoveredIndex] = useState(null);
    const navigate = useNavigate();

    const handleMouseEnter = (index) => {
        setHoveredIndex(index);
    };

    const handleMouseLeave = () => {
        setHoveredIndex(null);
    };

    const handleStrategySelect = (strategy) => {
        navigate(`/transit/${strategy.name}`, { state: { strategy } });
    };    

    const handleGoBack = () => {
        navigate(-1); // Go back to the previous page
    };

    return (
        <div style={styles.container}>
            <h1 style={styles.heading}>Select a Strategy</h1>
            <button onClick={handleGoBack} style={styles.backButton}>Back</button>
            <ul style={styles.strategyList}>
                {strategies.map((strategy, index) => (
                    <li key={index} style={styles.strategyItem}>
                        <button
                            onClick={() => handleStrategySelect(strategy)}
                            style={{
                                ...styles.strategyButton,
                                backgroundColor: hoveredIndex === index ? '#007bff' : 'transparent',
                                color: hoveredIndex === index ? '#fff' : '#007bff',
                            }}
                            onMouseEnter={() => handleMouseEnter(index)}
                            onMouseLeave={handleMouseLeave}
                        >
                            {strategy.name}
                        </button>
                    </li>
                ))}
            </ul>
        </div>
    );
};

export default StrategySelectionPage;

const styles = {
    container: {
        textAlign: 'center',
        marginTop: '50px',
        position: 'relative', // Add position relative
    },
    heading: {
        fontSize: '24px',
        marginBottom: '30px',
    },
    backButton: {
        backgroundColor: '#007bff',
        color: '#fff',
        border: 'none',
        borderRadius: '5px',
        padding: '8px 16px',
        cursor: 'pointer',
        position: 'absolute', // Change position to absolute
        left: '20px', // Set left distance
        top: '20px', // Set top distance
    },
    strategyList: {
        listStyle: 'none',
        padding: '0',
        margin: '0',
    },
    strategyItem: {
        marginBottom: '10px',
    },
    strategyButton: {
        textDecoration: 'none',
        color: '#007bff',
        fontSize: '18px',
        padding: '10px 20px',
        border: '1px solid #007bff',
        borderRadius: '5px',
        transition: 'background-color 0.3s, color 0.3s',
        display: 'inline-block',
        cursor: 'pointer',
    },
};
