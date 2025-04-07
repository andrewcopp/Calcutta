import { useState, useEffect } from 'react'
import { calcuttaService } from '../services/calcuttaService'
import { School } from '../types/calcutta'

export function SchoolList() {
    const [schools, setSchools] = useState<School[]>([])
    const [error, setError] = useState<string>('')
    const [loading, setLoading] = useState(true)
    const [searchTerm, setSearchTerm] = useState('')

    useEffect(() => {
        const loadSchools = async () => {
            try {
                setLoading(true)
                const data = await calcuttaService.getSchools()
                setSchools(data)
            } catch (err) {
                setError('Failed to load schools')
                console.error(err)
            } finally {
                setLoading(false)
            }
        }
        loadSchools()
    }, [])

    const filteredSchools = schools.filter(school => 
        school.name.toLowerCase().includes(searchTerm.toLowerCase())
    )

    if (error) return (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
            Error: {error}
        </div>
    )
    
    return (
        <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-2xl font-bold mb-4">Tournament Schools</h2>
            
            <div className="mb-4">
                <input
                    type="text"
                    placeholder="Search schools..."
                    className="w-full p-2 border border-gray-300 rounded-md"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                />
            </div>

            {loading ? (
                <div className="text-center py-4">Loading schools...</div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {filteredSchools.map(school => (
                        <div 
                            key={school.id}
                            className="p-3 border border-gray-200 rounded-md hover:bg-gray-50"
                        >
                            {school.name}
                        </div>
                    ))}
                </div>
            )}

            {!loading && filteredSchools.length === 0 && (
                <div className="text-center py-4 text-gray-500">
                    No schools found matching your search.
                </div>
            )}
        </div>
    )
}