import React, { useState, useEffect } from 'react';

const App = () => {
  const days = ['mon', 'tus', 'wed', 'thu', 'fri', 'sat', 'sun'];
  const times = ['nine_am', 'ten_am', 'eleven_am', 'twelve_pm', 'one_pm', 'two_pm', 'three_pm', 'four_pm'];
  const prettyTimes = ['9:00 AM', '10:00 AM', '11:00 AM', '12:00 PM', '1:00 PM', '2:00 PM', '3:00 PM', '4:00 PM'];

  const [schedule, setSchedule] = useState({});
  const [editing, setEditing] = useState(null);
  const [editValue, setEditValue] = useState('');

  useEffect(() => {
    const fetchSchedule = async () => {
        try {
            const response = await fetch('http://localhost:8080/schedule');
            if (!response.ok) {
                throw new Error('Failed to fetch schedule');
            }
            const data = await response.json();
            
            const transformedSchedule = {};
            data.forEach(item => {
                transformedSchedule[`${item.day}-nine_am`] = item.nine_am;
                transformedSchedule[`${item.day}-ten_am`] = item.ten_am;
                transformedSchedule[`${item.day}-eleven_am`] = item.eleven_am;
                transformedSchedule[`${item.day}-twelve_pm`] = item.twelve_pm;
                transformedSchedule[`${item.day}-one_pm`] = item.one_pm;
                transformedSchedule[`${item.day}-two_pm`] = item.two_pm;
                transformedSchedule[`${item.day}-three_pm`] = item.three_pm;
                transformedSchedule[`${item.day}-four_pm`] = item.four_pm;
            });
            
            setSchedule(transformedSchedule);
        } catch (error) {
            console.error('Error fetching schedule:', error);
        }
    };

    fetchSchedule();
  }, []);

  const handleCellClick = (day, time) => {
    setEditing({ day, time });
    setEditValue(schedule[`${day}-${time}`]?.slice(0, 30) || '');
  };

  const handleSave = async (day, time, task) => {
    try {
      const response = await fetch('http://localhost:8080/task', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ day, time, task }),
      });

      if (!response.ok) {
        throw new Error('Failed to save task');
      }

      setSchedule(prev => ({
        ...prev,
        [`${day}-${time}`]: task,
      }));
      setEditing(null);
      setEditValue('');
    } catch (error) {
      console.error('Error saving task:', error);
      alert('Failed to save task. Please try again.');
    }
  };

  const handleKeyPress = (e, day, time) => {
    if (e.key === 'Enter') {
      handleSave(day, time, editValue);
    } else if (e.key === 'Escape') {
      setEditing(null);
      setEditValue('');
    }
  };

  const handleReset = async () => {
    try {
      const response = await fetch('http://localhost:8080/reset', {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error('Failed to reset table');
      }

      setSchedule({}); // Clear the local state
    } catch (error) {
      console.error('Error resetting table:', error);
      alert('Failed to reset table. Please try again.');
    }
  };

  return (
    <div style={containerStyle}>
      <h1 style={titleStyle}>Weekly Schedule</h1>
      <div style={{ overflowX: 'auto' }}>
        <table style={tableStyle}>
          <thead>
            <tr>
              <th style={headerStyle}>Day / Time</th>
              {times.map((time, index) => (
                <th key={time} style={headerStyle}>
                  {prettyTimes[index]}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {days.map(day => (
              <tr key={day}>
                <td style={headerStyle}>{day.toUpperCase()}</td>
                {times.map(time => (
                  <td
                    key={`${day}-${time}`}
                    style={cellStyle}
                    onClick={() => handleCellClick(day, time)}
                  >
                    {editing?.day === day && editing?.time === time ? (
                      <input
                        type="text"
                        style={inputStyle}
                        value={editValue}
                        onChange={(e) =>
                          setEditValue(e.target.value.slice(0, 30))
                        }
                        maxLength={30}
                        onKeyDown={(e) => handleKeyPress(e, day, time)}
                        autoFocus
                      />
                    ) : (
                      <div
                        style={{
                          minHeight: '40px',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                        }}
                      >
                        {schedule[`${day}-${time}`] || ''}
                      </div>
                    )}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <button 
        onClick={handleReset}
        style={resetButtonStyle}
      >
        Reset Table
      </button>
    </div>
  );
};

// Enhanced styling objects for a modern, professional look

const containerStyle = {
  padding: '30px',
  maxWidth: '1200px',
  margin: '40px auto',
  backgroundColor: '#f4f7f9',
  borderRadius: '8px',
  fontFamily: '"Segoe UI", Tahoma, Geneva, Verdana, sans-serif',
};

const titleStyle = {
  textAlign: 'center',
  marginBottom: '20px',
  color: '#2c3e50',
  fontSize: '28px',
};

const tableStyle = {
  width: '100%',
  borderCollapse: 'collapse',
  backgroundColor: '#fff',
  boxShadow: '0 4px 15px rgba(0, 0, 0, 0.1)',
  borderRadius: '8px',
  overflow: 'hidden',
};

const headerStyle = {
  padding: '16px',
  border: '1px solid #e0e0e0',
  backgroundColor: '#34495e',
  color: '#ecf0f1',
  fontWeight: '600',
  textAlign: 'center',
  fontSize: '16px',
  textTransform: 'uppercase',
};

const cellStyle = {
  padding: '12px',
  border: '1px solid #ecf0f1',
  cursor: 'pointer',
  position: 'relative',
  verticalAlign: 'middle',
  textAlign: 'center',
  fontSize: '15px',
  color: '#2c3e50',
};

const inputStyle = {
  width: '100%',
  height: '100%',
  padding: '10px',
  border: '2px solid #2980b9',
  borderRadius: '4px',
  outline: 'none',
  boxSizing: 'border-box',
};

const resetButtonStyle = {
  marginTop: '20px',
  padding: '10px 20px',
  backgroundColor: '#e74c3c',
  color: 'white',
  border: 'none',
  borderRadius: '4px',
  cursor: 'pointer',
  fontSize: '16px',
  fontWeight: '600',
  display: 'block',
  margin: '20px auto 0',
  transition: 'background-color 0.2s',
  ':hover': {
    backgroundColor: '#c0392b',
  }
};

export default App;
