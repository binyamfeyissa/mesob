import { KYCReviewList } from "@/features/kyc-review/KYCReviewList";

export default function KYCReviewPage() {
  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold text-gray-900 mb-6">KYC Review</h1>
      <KYCReviewList />
    </div>
  );
}
