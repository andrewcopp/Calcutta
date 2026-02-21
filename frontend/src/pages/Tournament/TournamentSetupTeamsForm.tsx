import type { Tournament, TournamentTeam } from '../../types/tournament';
import type { School } from '../../types/school';
import { useTeamSetupForm } from '../../hooks/useTeamSetupForm';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../../components/ui/Tabs';
import { ValidationPanel } from './ValidationPanel';
import { RegionSeedForm } from './RegionSeedForm';

interface TournamentSetupTeamsFormProps {
  tournament: Tournament;
  schools: School[];
  schoolOptions: { id: string; label: string }[];
  initialTeams?: TournamentTeam[];
}

export function TournamentSetupTeamsForm({
  tournament,
  schools,
  schoolOptions,
  initialTeams,
}: TournamentSetupTeamsFormProps) {
  const {
    regionList,
    regions,
    activeTab,
    setActiveTab,
    errors,
    stats,
    slotValidation,
    usedSchoolIds,
    updateSlot,
    addPlayIn,
    removePlayIn,
    handleSlotBlur,
    handleSubmit,
    isPending,
    createEmptyRegion,
  } = useTeamSetupForm({ tournament, schools, initialTeams });

  return (
    <>
      {errors.length > 0 && (
        <Alert variant="error" className="mb-4">
          <div className="font-semibold mb-1">Validation Errors</div>
          <ul className="list-disc list-inside text-sm space-y-1">
            {errors.map((err, i) => (
              <li key={i}>{err}</li>
            ))}
          </ul>
        </Alert>
      )}

      <ValidationPanel stats={stats} regionNames={regionList} />

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          {regionList.map((regionName) => (
            <TabsTrigger key={regionName} value={regionName}>
              {regionName} ({stats.perRegion[regionName] || 0})
            </TabsTrigger>
          ))}
        </TabsList>

        {regionList.map((regionName) => (
          <TabsContent key={regionName} value={regionName}>
            <Card>
              <RegionSeedForm
                region={regionName}
                regionState={regions[regionName] || createEmptyRegion()}
                schoolOptions={schoolOptions}
                schools={schools}
                usedSchoolIds={usedSchoolIds}
                updateSlot={updateSlot}
                addPlayIn={addPlayIn}
                removePlayIn={removePlayIn}
                onSlotBlur={handleSlotBlur}
                slotValidation={slotValidation[regionName]}
              />
            </Card>
          </TabsContent>
        ))}
      </Tabs>

      <div className="flex justify-end mt-6">
        <Button
          onClick={handleSubmit}
          disabled={isPending || stats.total === 0}
          loading={isPending}
        >
          {isPending ? 'Saving...' : 'Save Teams'}
        </Button>
      </div>
    </>
  );
}
