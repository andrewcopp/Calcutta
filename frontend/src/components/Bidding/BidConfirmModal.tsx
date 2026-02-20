import { Modal, ModalActions } from '../ui/Modal';
import { Button } from '../ui/Button';
import { Badge } from '../ui/Badge';
import { getSeedVariant } from '../../hooks/useBidding';
import type { PortfolioItem } from '../../hooks/useBidding';

interface BidConfirmModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isPending: boolean;
  portfolioSummary: PortfolioItem[];
  totalBudget: number;
  budgetRemaining: number;
}

export function BidConfirmModal({
  open,
  onClose,
  onConfirm,
  isPending,
  portfolioSummary,
  totalBudget,
  budgetRemaining,
}: BidConfirmModalProps) {
  return (
    <Modal open={open} onClose={onClose} title="Confirm Your Bids">
      <div className="space-y-4">
        <div className="flex justify-between text-sm text-gray-600">
          <span>{portfolioSummary.length} teams selected</span>
          <span>Total spent: {totalBudget - budgetRemaining} pts</span>
        </div>
        <div className="text-sm text-gray-600">
          Budget remaining: {budgetRemaining} pts
        </div>
        <div className="max-h-60 overflow-y-auto space-y-2">
          {portfolioSummary.map((item) => (
            <div key={item.teamId} className="flex items-center justify-between py-1 border-b border-gray-100">
              <div className="flex items-center gap-2">
                <Badge variant={getSeedVariant(item.seed)} className="text-xs">{item.seed}</Badge>
                <span className="text-sm text-gray-800">{item.name}</span>
              </div>
              <span className="text-sm font-medium text-blue-700">{item.bid} pts</span>
            </div>
          ))}
        </div>
      </div>
      <ModalActions>
        <Button variant="secondary" onClick={onClose}>Go Back</Button>
        <Button onClick={onConfirm} loading={isPending}>Confirm &amp; Submit</Button>
      </ModalActions>
    </Modal>
  );
}
