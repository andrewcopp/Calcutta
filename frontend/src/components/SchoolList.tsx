import { useState, useEffect } from 'react'

interface School {
    id: string
    name: string
}

export function SchoolList() {
    const [schools, setSchools] = useState<School[]>([])
    const [error, setError] = useState<string>('')

    useEffect(() => {
        fetch('http://localhost:8080/api/schools')
            .then(response => response.json())
            .then(data => setSchools(data))
            .catch(err => setError('Failed to load schools'))
    }, [])

    if (error) return <div>Error: {error}</div>
    
    return (
        <div>
            <h2>Tournament Schools</h2>
            <ul>
                {schools.map(school => (
                    <li key={school.id}>{school.name}</li>
                ))}
            </ul>
        </div>
    )
}