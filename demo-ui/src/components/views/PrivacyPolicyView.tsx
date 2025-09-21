import React from "react";

/**
 * Privacy Policy View for RecSys Demo UI
 *
 * This component displays the privacy policy for the recommendation system API,
 * covering data collection, usage, storage, and user rights.
 */
export function PrivacyPolicyView() {
  return (
    <div style={{ maxWidth: 800, margin: "0 auto", padding: "20px" }}>
      <h1 style={{ color: "#1976d2", marginBottom: 24 }}>Privacy Policy</h1>

      <div
        style={{
          backgroundColor: "#f8f9fa",
          padding: 16,
          borderRadius: 8,
          marginBottom: 24,
          border: "1px solid #e9ecef",
        }}
      >
        <p style={{ margin: "8px 0 0 0", fontSize: 14, color: "#6c757d" }}>
          This privacy policy describes how the RecSys Demo API collects, uses,
          and protects your information when using our recommendation system.
        </p>
      </div>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>
          1. Information We Collect
        </h2>

        <h3 style={{ color: "#555", marginBottom: 12 }}>User Data</h3>
        <ul style={{ marginBottom: 16, paddingLeft: 20 }}>
          <li>
            <strong>User Identifiers:</strong> Unique user IDs for
            personalization
          </li>
          <li>
            <strong>User Traits:</strong> Demographic and preference data (age,
            location, interests)
          </li>
          <li>
            <strong>Behavioral Data:</strong> User interactions with items and
            content
          </li>
        </ul>

        <h3 style={{ color: "#555", marginBottom: 12 }}>Item Data</h3>
        <ul style={{ marginBottom: 16, paddingLeft: 20 }}>
          <li>
            <strong>Item Identifiers:</strong> Unique item IDs and metadata
          </li>
          <li>
            <strong>Item Properties:</strong> Categories, tags, prices, and
            descriptions
          </li>
          <li>
            <strong>Embeddings:</strong> Vector representations for similarity
            matching
          </li>
        </ul>

        <h3 style={{ color: "#555", marginBottom: 12 }}>Interaction Data</h3>
        <ul style={{ marginBottom: 16, paddingLeft: 20 }}>
          <li>
            <strong>Event Types:</strong> Views, clicks, cart additions,
            purchases
          </li>
          <li>
            <strong>Timestamps:</strong> When interactions occurred
          </li>
          <li>
            <strong>Context:</strong> Additional metadata about interactions
          </li>
        </ul>
      </section>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>
          2. How We Use Your Information
        </h2>

        <ul style={{ paddingLeft: 20 }}>
          <li>
            <strong>Personalization:</strong> Generate personalized
            recommendations
          </li>
          <li>
            <strong>Analytics:</strong> Understand user behavior and system
            performance
          </li>
          <li>
            <strong>Algorithm Improvement:</strong> Train and optimize
            recommendation models
          </li>
          <li>
            <strong>A/B Testing:</strong> Test different recommendation
            strategies
          </li>
          <li>
            <strong>Bandit Optimization:</strong> Continuously improve
            recommendation quality
          </li>
        </ul>
      </section>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>
          3. Data Storage and Security
        </h2>

        <h3 style={{ color: "#555", marginBottom: 12 }}>Storage</h3>
        <ul style={{ marginBottom: 16, paddingLeft: 20 }}>
          <li>Data is stored in secure, encrypted databases</li>
          <li>User data is organized by organization and namespace</li>
          <li>Data retention follows organizational policies</li>
        </ul>

        <h3 style={{ color: "#555", marginBottom: 12 }}>Security Measures</h3>
        <ul style={{ marginBottom: 16, paddingLeft: 20 }}>
          <li>Encryption in transit and at rest</li>
          <li>Access controls and authentication</li>
          <li>Regular security audits and monitoring</li>
          <li>Data anonymization where possible</li>
        </ul>
      </section>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>
          4. Data Sharing and Third Parties
        </h2>

        <p style={{ marginBottom: 16 }}>
          We do not sell, trade, or rent your personal information to third
          parties. Data may be shared only in the following circumstances:
        </p>

        <ul style={{ paddingLeft: 20 }}>
          <li>With your explicit consent</li>
          <li>To comply with legal obligations</li>
          <li>
            With service providers under strict confidentiality agreements
          </li>
          <li>In aggregated, anonymized form for research purposes</li>
        </ul>
      </section>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>
          5. Your Rights and Choices
        </h2>

        <h3 style={{ color: "#555", marginBottom: 12 }}>
          Data Access and Control
        </h3>
        <ul style={{ marginBottom: 16, paddingLeft: 20 }}>
          <li>
            <strong>Access:</strong> Request a copy of your personal data
          </li>
          <li>
            <strong>Correction:</strong> Update or correct inaccurate
            information
          </li>
          <li>
            <strong>Deletion:</strong> Request deletion of your personal data
          </li>
          <li>
            <strong>Portability:</strong> Export your data in a machine-readable
            format
          </li>
        </ul>

        <h3 style={{ color: "#555", marginBottom: 12 }}>Opt-out Options</h3>
        <ul style={{ marginBottom: 16, paddingLeft: 20 }}>
          <li>Disable personalized recommendations</li>
          <li>Opt out of data collection for analytics</li>
          <li>Request data anonymization</li>
        </ul>
      </section>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>
          6. Cookies and Tracking
        </h2>

        <p style={{ marginBottom: 16 }}>
          This demo application may use cookies and similar technologies to:
        </p>

        <ul style={{ paddingLeft: 20 }}>
          <li>Maintain user session state</li>
          <li>Remember user preferences</li>
          <li>Track API usage for analytics</li>
          <li>Improve user experience</li>
        </ul>

        <p style={{ marginTop: 16 }}>
          You can control cookie settings through your browser preferences.
        </p>
      </section>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>7. Data Retention</h2>

        <p style={{ marginBottom: 16 }}>
          We retain your data for as long as necessary to provide our services
          and comply with legal obligations. Specific retention periods include:
        </p>

        <ul style={{ paddingLeft: 20 }}>
          <li>
            <strong>User Profiles:</strong> Until account deletion or 3 years of
            inactivity
          </li>
          <li>
            <strong>Interaction Events:</strong> 2 years for analytics and model
            training
          </li>
          <li>
            <strong>Item Data:</strong> Until items are no longer available
          </li>
          <li>
            <strong>System Logs:</strong> 90 days for security and debugging
          </li>
        </ul>
      </section>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>
          8. International Data Transfers
        </h2>

        <p>
          Your data may be transferred to and processed in countries other than
          your own. We ensure appropriate safeguards are in place to protect
          your data in accordance with applicable privacy laws.
        </p>
      </section>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>
          9. Children's Privacy
        </h2>

        <p>
          Our services are not intended for children under 13 years of age. We
          do not knowingly collect personal information from children under 13.
          If we become aware of such collection, we will take steps to delete
          the information promptly.
        </p>
      </section>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>
          10. Changes to This Policy
        </h2>

        <p>
          We may update this privacy policy from time to time. Your continued
          use of our services after such changes constitutes acceptance of the
          updated policy.
        </p>
      </section>

      <section style={{ marginBottom: 32 }}>
        <h2 style={{ color: "#333", marginBottom: 16 }}>
          11. Contact Information
        </h2>

        <p style={{ marginBottom: 16 }}>
          If you have any questions about this privacy policy or our data
          practices, please contact us:
        </p>

        <div
          style={{
            backgroundColor: "#f8f9fa",
            padding: 16,
            borderRadius: 8,
            border: "1px solid #e9ecef",
          }}
        >
          <p style={{ margin: "0 0 8px 0" }}>
            <strong>Email:</strong> privacy@recsys-demo.com
          </p>
          <p style={{ margin: "0 0 8px 0" }}>
            <strong>Data Protection Officer:</strong> dpo@recsys-demo.com
          </p>
          <p style={{ margin: 0 }}>
            <strong>Response Time:</strong> We will respond to privacy inquiries
            within 30 days.
          </p>
        </div>
      </section>

      <div
        style={{
          marginTop: 40,
          padding: 20,
          backgroundColor: "#e3f2fd",
          borderRadius: 8,
          border: "1px solid #bbdefb",
        }}
      >
        <h3 style={{ color: "#1976d2", marginBottom: 12 }}>
          Demo Environment Notice
        </h3>
        <p style={{ margin: 0, fontSize: 14, color: "#555" }}>
          <strong>Important:</strong> This is a demonstration environment. The
          data you interact with here is for testing purposes only and may be
          reset or modified at any time. Please do not enter real personal
          information in this demo system.
        </p>
      </div>
    </div>
  );
}
